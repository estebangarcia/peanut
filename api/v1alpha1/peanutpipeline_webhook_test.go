package v1alpha1

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("PeanutPipeline webhook", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	AfterEach(func() {
		stage := &PeanutStage{}

		k8sClient.Get(context.TODO(), types.NamespacedName{
			Name:      "stage-test",
			Namespace: "default",
		}, stage)

		k8sClient.Delete(context.TODO(), stage)
	})

	Context("PeanutPipeline validation", func() {
		It("should fail to validate due to stages not found", func() {
			pipeline := &PeanutPipeline{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pipeline-test",
					Namespace: "default",
				},
				Spec: PeanutPipelineSpec{
					Stages: []PeanutPipelineStageDefinition{
						{Name: "First Stage", Type: "stage-one", Dependencies: nil},
						{Name: "Second Stage", Type: "stage-two", Dependencies: nil},
						{Name: "Third Stage", Type: "stage-three", Dependencies: []string{"Second Stage"}},
						{Name: "Fourth Stage", Type: "stage-four", Dependencies: []string{"Third Stage"}},
						{Name: "Fifth Stage", Type: "stage-five", Dependencies: []string{"First Stage", "Second Stage"}},
						{Name: "Sixth Stage", Type: "stage-six", Dependencies: []string{"First Stage"}},
					},
				},
				Status: PeanutPipelineStatus{},
			}

			err := pipeline.validatePipeline()
			Expect(err).ToNot(BeNil())
		})

		It("should succeed", func() {
			stageToCreate := &PeanutStage{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "stage-test",
					Namespace: "default",
				},
				Spec: PeanutStageSpec{
					Image:   "testing",
					Timeout: 30,
				},
			}

			By("Creating the PeanutStage successfully")
			Expect(createStage(stageToCreate)).Should(Succeed())

			stage := &PeanutStage{}
			Eventually(func() bool {
				k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: "default",
					Name:      "stage-test",
				}, stage)
				return stage != nil && stage.Name == "stage-test"
			}, timeout, interval).Should(BeTrue())

			pipeline := &PeanutPipeline{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pipeline-test",
					Namespace: "default",
				},
				Spec: PeanutPipelineSpec{
					Stages: []PeanutPipelineStageDefinition{
						{Name: "First Stage", Type: "stage-test", Dependencies: nil},
					},
				},
			}

			err := pipeline.validatePipeline()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail because the stage is duplicated", func() {
			stageToCreate := &PeanutStage{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "stage-test",
					Namespace: "default",
				},
				Spec: PeanutStageSpec{
					Image:   "testing",
					Timeout: 30,
				},
			}

			By("Creating the PeanutStage successfully")
			Expect(createStage(stageToCreate)).Should(Succeed())

			pipeline := &PeanutPipeline{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pipeline-test",
					Namespace: "default",
				},
				Spec: PeanutPipelineSpec{
					Stages: []PeanutPipelineStageDefinition{
						{Name: "First Stage", Type: "stage-test", Dependencies: nil},
						{Name: "First Stage", Type: "stage-test", Dependencies: nil},
					},
				},
				Status: PeanutPipelineStatus{},
			}

			err := pipeline.validatePipeline()
			Expect(err).ToNot(BeNil())
		})
	})
})

func createStage(stage *PeanutStage) error {
	if err := k8sClient.Create(context.TODO(), stage); err != nil {
		return err
	}
	return nil
}
