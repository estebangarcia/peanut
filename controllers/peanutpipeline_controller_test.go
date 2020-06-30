package controllers

import (
	"context"
	"github.com/estebangarcia/peanut/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("PeanutPipeline Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	key := types.NamespacedName{
		Name:      "pipeline-test",
		Namespace: "default",
	}

	AfterEach(func() {
		pipeline := &v1alpha1.PeanutPipeline{}
		k8sClient.Get(context.TODO(), key, pipeline)
		k8sClient.Delete(context.TODO(), pipeline)
	})

	Context("PeanutPipeline creation", func() {
		It("Should create PeanutPipeline successfully and assign status", func() {
			pipelineToCreate := &v1alpha1.PeanutPipeline{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: v1alpha1.PeanutPipelineSpec{
					Stages: []v1alpha1.PeanutPipelineStageDefinition{
						{Name: "First Stage", Type: "stage-one", Dependencies: nil},
						{Name: "Second Stage", Type: "stage-two", Dependencies: nil},
						{Name: "Third Stage", Type: "stage-three", Dependencies: []string{"Second Stage"}},
						{Name: "Fourth Stage", Type: "stage-four", Dependencies: []string{"Third Stage"}},
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
		})
	})
})
