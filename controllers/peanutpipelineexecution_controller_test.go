package controllers

import (
	"context"
	"github.com/estebangarcia/peanut/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

var _ = Describe("PeanutPipelineExecution Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	AfterEach(func() {
		pipeline := &v1alpha1.PeanutPipeline{}
		stage := &v1alpha1.PeanutStage{}
		execution := &v1alpha1.PeanutPipelineExecution{}

		k8sClient.Get(context.TODO(), types.NamespacedName{
			Name:      "stage-test",
			Namespace: "default",
		}, stage)
		k8sClient.Get(context.TODO(), types.NamespacedName{
			Name:      "pipeline-test",
			Namespace: "default",
		}, pipeline)
		k8sClient.Get(context.TODO(), types.NamespacedName{
			Name:      "pipelineexecution-test",
			Namespace: "default",
		}, execution)

		k8sClient.Delete(context.TODO(), stage)
		k8sClient.Delete(context.TODO(), pipeline)
		k8sClient.Delete(context.TODO(), execution)
	})

	Context("PeanutPipelineExecution creation", func() {
		It("Should Add ReferenceOwner to PeanutPipeline", func() {

			pipeline := createResources(timeout, interval)

			executionKey := types.NamespacedName{
				Name:      "pipelineexecution-test",
				Namespace: "default",
			}

			pipelineExecutionToCreate := &v1alpha1.PeanutPipelineExecution{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      executionKey.Name,
					Namespace: executionKey.Namespace,
				},
				Spec: v1alpha1.PeanutPipelineExecutionSpec{
					PipelineName: "pipeline-test",
				},
			}

			By("Creating the PeanutPipelineExecution successfully")
			Expect(k8sClient.Create(context.TODO(), pipelineExecutionToCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			pipelineExecution := &v1alpha1.PeanutPipelineExecution{}
			Eventually(func() bool {
				k8sClient.Get(context.TODO(), executionKey, pipelineExecution)
				return pipelineExecution != nil && pipelineExecution.OwnerReferences[0].UID == pipeline.UID
			}, timeout, interval).Should(BeTrue())
		})

		It("Should create jobs by traversing stage graph", func() {
			createResources(timeout, interval)

			executionKey := types.NamespacedName{
				Name:      "pipelineexecution-test",
				Namespace: "default",
			}

			pipelineExecutionToCreate := &v1alpha1.PeanutPipelineExecution{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      executionKey.Name,
					Namespace: executionKey.Namespace,
				},
				Spec: v1alpha1.PeanutPipelineExecutionSpec{
					PipelineName: "pipeline-test",
				},
			}

			By("Creating the PeanutPipelineExecution successfully")
			Expect(k8sClient.Create(context.TODO(), pipelineExecutionToCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)

			jobList := &v1.JobList{}
			Eventually(func() bool {
				k8sClient.List(context.TODO(), jobList, client.InNamespace("default"))
				return len(jobList.Items) == 2
			}, timeout, interval).Should(BeTrue())

			jobStage := &v1.Job{}

			jobNames := flattenJobList(jobList)
			jobStage = findJobInList(jobList, "pipeline-test-secondstage-")

			Expect(jobNames).To(ContainElements(
				ContainSubstring("pipeline-test-firststage-"),
				ContainSubstring("pipeline-test-secondstage-"),
			))

			// Update condition of the job to mock completion
			jobStage.Status.Conditions = append(jobStage.Status.Conditions, v1.JobCondition{
				Type:   v1.JobComplete,
				Status: "True",
				LastProbeTime: metav1.Time{
					Time: time.Now(),
				},
				LastTransitionTime: metav1.Time{
					Time: time.Now(),
				},
			})

			Expect(k8sClient.Status().Update(context.TODO(), jobStage)).Should(Succeed())

			Eventually(func() bool {
				k8sClient.List(context.TODO(), jobList, client.InNamespace("default"))
				return len(jobList.Items) == 3
			}, timeout, interval).Should(BeTrue())

			jobNames = flattenJobList(jobList)
			jobStage = findJobInList(jobList, "pipeline-test-thirdstage-")

			Expect(jobNames).To(ContainElements(
				ContainSubstring("pipeline-test-firststage-"),
				ContainSubstring("pipeline-test-secondstage-"),
				ContainSubstring("pipeline-test-thirdstage-"),
			))
			// Update condition of the job to mock completion
			jobStage.Status.Conditions = append(jobStage.Status.Conditions, v1.JobCondition{
				Type:   v1.JobComplete,
				Status: "True",
				LastProbeTime: metav1.Time{
					Time: time.Now(),
				},
				LastTransitionTime: metav1.Time{
					Time: time.Now(),
				},
			})

			Expect(k8sClient.Status().Update(context.TODO(), jobStage)).Should(Succeed())

			Eventually(func() bool {
				k8sClient.List(context.TODO(), jobList, client.InNamespace("default"))
				return len(jobList.Items) == 4
			}, timeout, interval).Should(BeTrue())

			jobNames = flattenJobList(jobList)

			Expect(jobNames).To(ContainElements(
				ContainSubstring("pipeline-test-firststage-"),
				ContainSubstring("pipeline-test-secondstage-"),
				ContainSubstring("pipeline-test-thirdstage-"),
				ContainSubstring("pipeline-test-fourthstage-"),
			))

		})
	})
})

func flattenJobList(list *v1.JobList) []string {
	jobNames := []string{}
	for _, job := range list.Items {
		jobNames = append(jobNames, job.Name)
	}

	return jobNames
}

func findJobInList(list *v1.JobList, name string) *v1.Job {
	for _, job := range list.Items {
		if strings.Contains(job.Name, name) {
			return &job
		}
	}

	return nil
}

func createResources(timeout time.Duration, interval time.Duration) *v1alpha1.PeanutPipeline {
	stageToCreate := &v1alpha1.PeanutStage{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "stage-test",
			Namespace: "default",
		},
		Spec: v1alpha1.PeanutStageSpec{
			Image:   "testing",
			Timeout: 30,
		},
	}

	By("Creating the PeanutStage successfully")
	Expect(k8sClient.Create(context.TODO(), stageToCreate)).Should(Succeed())
	time.Sleep(time.Second * 5)

	key := types.NamespacedName{
		Name:      "pipeline-test",
		Namespace: "default",
	}

	pipelineToCreate := &v1alpha1.PeanutPipeline{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: v1alpha1.PeanutPipelineSpec{
			Stages: []v1alpha1.PeanutPipelineStageDefinition{
				{Name: "First Stage", Type: "stage-test", Dependencies: nil},
				{Name: "Second Stage", Type: "stage-test", Dependencies: nil},
				{Name: "Third Stage", Type: "stage-test", Dependencies: []string{"Second Stage"}},
				{Name: "Fourth Stage", Type: "stage-test", Dependencies: []string{"Third Stage"}},
			},
		},
	}

	By("Creating the PeanutPipeline successfully")
	Expect(k8sClient.Create(context.TODO(), pipelineToCreate)).Should(Succeed())
	time.Sleep(time.Second * 5)

	pipeline := &v1alpha1.PeanutPipeline{}
	Eventually(func() bool {
		k8sClient.Get(context.TODO(), key, pipeline)
		return pipeline != nil && pipeline.Status.Ack == "true"
	}, timeout, interval).Should(BeTrue())

	return pipeline
}
