package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PeanutPipeline", func() {
	Describe("Build stage graph", func() {
		It("should create stage graph", func() {
			pipeline := PeanutPipeline{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
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

			graph, err := pipeline.BuildStageGraph()
			Expect(err).ToNot(HaveOccurred())
			Expect(graph).ToNot(BeNil())
			Expect(graph.Size()).To(BeEquivalentTo(7))
			Expect(graph.Order()).To(BeEquivalentTo(7))

			root, err := graph.GetVertex("root")
			Expect(err).ToNot(HaveOccurred())
			Expect(root).ToNot(BeNil())

			type graphT map[string][]string
			values := graphT{
				"root": []string{
					"First Stage",
					"Second Stage",
				},
				"First Stage": []string{
					"Fifth Stage",
					"Sixth Stage",
				},
				"Second Stage": []string{
					"Third Stage",
					"Fifth Stage",
				},
				"Third Stage": []string{
					"Fourth Stage",
				},
				"Fourth Stage": []string{},
				"Fifth Stage":  []string{},
				"Sixth Stage":  []string{},
			}

			for id, children := range values {
				v, err := graph.GetVertex(id)
				Expect(err).ToNot(HaveOccurred())

				Expect(v.Children.Size()).To(BeEquivalentTo(len(children)))

				for _, c := range children {
					vC, err := graph.GetVertex(c)
					Expect(err).ToNot(HaveOccurred())

					Expect(vC.Parents.Contains(v)).To(BeTrue())
				}
			}

		})
	})
})
