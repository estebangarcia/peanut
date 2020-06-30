/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/goombaio/dag"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilpointer "k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strings"
	"time"

	peanut "github.com/estebangarcia/peanut/api/v1alpha1"
)

var (
	jobOwnerKey             = ".metadata.controller"
	apiGVStr                = peanut.GroupVersion.String()
	scheduledTimeAnnotation = "pipeline.peanut.dev/scheduled-at"
	stageNameAnnotation     = "pipeline.peanut.dev/name"
)

// PeanutPipelineExecutionReconciler reconciles a PeanutPipelineExecution object
type PeanutPipelineExecutionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=pipeline.peanut.dev,resources=peanutpipelineexecutions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pipeline.peanut.dev,resources=peanutpipelineexecutions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

func (r *PeanutPipelineExecutionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("peanutpipelineexecution", req.NamespacedName)

	var pipelineExecution peanut.PeanutPipelineExecution
	if err := r.Get(ctx, req.NamespacedName, &pipelineExecution); err != nil {
		log.Error(err, "unable to fetch PeanutPipelineExecution")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var pipelineDefinition peanut.PeanutPipeline
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      pipelineExecution.Spec.PipelineName,
	}, &pipelineDefinition); err != nil {
		log.Error(err, "unable to fetch PeanutPipeline", "name", pipelineExecution.Spec.PipelineName)
	}

	if len(pipelineExecution.OwnerReferences) == 0 {
		pipelineExecution.OwnerReferences = append(pipelineExecution.OwnerReferences, metav1.OwnerReference{
			APIVersion:         apiGVStr,
			Kind:               pipelineDefinition.Kind,
			Name:               pipelineDefinition.Name,
			UID:                pipelineDefinition.UID,
			Controller:         utilpointer.BoolPtr(true),
			BlockOwnerDeletion: utilpointer.BoolPtr(true),
		})
		if err := r.Update(ctx, &pipelineExecution); err != nil {
			log.Error(err, "unable to update PipelineExecution OwnerReference")
			return ctrl.Result{}, err
		}
	}

	stageGraph, err := pipelineDefinition.BuildStageGraph()
	if err != nil {
		log.Error(err, "unable to build graph for PeanutPipeline", "name", pipelineDefinition.Name)
	}

	var childJobs kbatch.JobList
	if err := r.List(ctx, &childJobs, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err != nil {
		log.Error(err, "unable to list child Job")
		return ctrl.Result{}, err
	}

	childJobsItems := childJobs.Items

	var nextBatch []peanut.PeanutPipelineStageDefinition
	markPipelineAsFailed := false

	sort.SliceStable(childJobsItems, func(i, j int) bool {
		timeA, _ := time.Parse(time.RFC3339, childJobsItems[i].Annotations[scheduledTimeAnnotation])
		timeB, _ := time.Parse(time.RFC3339, childJobsItems[j].Annotations[scheduledTimeAnnotation])
		return timeA.Before(timeB)
	})

	if len(childJobsItems) > 0 {
		for _, job := range childJobsItems {
			// Get vertex from graph that corresponds to the current job in the iteration
			v, err := stageGraph.GetVertex(job.Annotations[stageNameAnnotation])
			if err != nil {
				log.Error(err, "job is controlled by pipelineexecution but stage is not in graph", "job", job, "graph", stageGraph)
				return ctrl.Result{}, err
			}

			// We get current job status
			_, finishedType := isJobFinished(&job)

			// If failed then we should mark the pipeline as failed
			// TODO: Decide if we keep schedule jobs for branches that are not dependent of the failed job, fail entire pipeline or let the user decide
			if finishedType == kbatch.JobFailed {
				markPipelineAsFailed = true
			} else if finishedType == kbatch.JobComplete {
				// If Job is complete we check for dependent stages and schedule them if needed
				for _, child := range v.Children.Values() {
					stage := child.(*dag.Vertex).Value.(peanut.PeanutPipelineStageDefinition)

					var foundJob *kbatch.Job
					for _, job := range childJobsItems {
						if job.Annotations[stageNameAnnotation] == stage.Name {
							foundJob = &job
						}
					}

					// Only do the following if the job wasn't already scheduled
					if foundJob == nil {
						canSchedule := true

						// We need to check if the current children don't depend of other jobs that might have
						// failed, are running or are still not scheduled
						childV, _ := stageGraph.GetVertex(stage.Name)
						for _, parentChildV := range childV.Parents.Values() {
							stageDef := parentChildV.(*dag.Vertex).Value.(peanut.PeanutPipelineStageDefinition)

							// Try to find if the parent was scheduled and its current status
							for _, job := range childJobsItems {
								if job.Annotations[stageNameAnnotation] == stageDef.Name {
									foundJob = &job
								}
							}

							// If job was found and it was successful then we schedule
							if foundJob != nil {
								_, finishedType := isJobFinished(foundJob)
								canSchedule = finishedType == kbatch.JobComplete
							} else {
								// Parent Job is not scheduled yet, then we don't schedule its child
								canSchedule = false
							}

						}

						if canSchedule {
							// Check if the job wasn't already scheduled because of multiple stages
							// with the same children
							found := false
							for _, s := range nextBatch {
								if stage.Name == s.Name {
									found = true
									break
								}
							}
							if !found {
								nextBatch = append(nextBatch, stage)
							}
						}
					}
				}

			}
		}
	} else {
		// If there are no jobs created then we add to the nextBatch the children of the root vertex
		rootVertex, err := stageGraph.GetVertex("root")
		if err != nil {
			log.Error(err, "no root vertex found")
			return ctrl.Result{}, err
		}

		vertices, err := stageGraph.Successors(rootVertex)
		if err != nil {
			log.Error(err, "no successors to the root vertex found?")
			return ctrl.Result{}, err
		}

		for _, v := range vertices {
			pipelineStageDefinition := v.Value.(peanut.PeanutPipelineStageDefinition)
			nextBatch = append(nextBatch, pipelineStageDefinition)
		}
	}

	if markPipelineAsFailed {
		pipelineExecution.Status.Phase = peanut.Failed
		if err := r.Status().Update(ctx, &pipelineExecution); err != nil {
			log.Error(err, "unable to update PipelineExecution status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	for _, stage := range nextBatch {
		var stageDefinition peanut.PeanutStage
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: req.Namespace,
			Name:      stage.Type,
		}, &stageDefinition); err != nil {
			log.Error(err, "unable to fetch PeanutStage with name %s", "name", stage.Type)
		}

		job := createJobForStage(pipelineDefinition.Name, stage.Name, req.Namespace, &stageDefinition, time.Now())

		if err := ctrl.SetControllerReference(&pipelineExecution, job, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, job); err != nil {
			log.Error(err, "unable to create Job for PipelineExecution", "job", job)
			return ctrl.Result{}, err
		}

		log.V(1).Info("created Job for PipelineExecution", "job", job)
	}

	return ctrl.Result{}, nil
}

func (r *PeanutPipelineExecutionReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &kbatch.Job{}, jobOwnerKey, func(rawObj runtime.Object) []string {
		// grab the job object, extract the owner...
		job := rawObj.(*kbatch.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}

		if owner.APIVersion != apiGVStr || owner.Kind != "PeanutPipelineExecution" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&peanut.PeanutPipelineExecution{}).
		Owns(&kbatch.Job{}).
		Complete(r)
}

func createJobForStage(pipelineName string, pipelineStageName string, namespace string, peanutStage *peanut.PeanutStage, scheduledTime time.Time) *kbatch.Job {
	// TODO: We want job to have a deterministic name to avoid the same job being created twice. Improve!
	sName := strings.ReplaceAll(pipelineStageName, " ", "")

	name := fmt.Sprintf("%s-%s-%d", pipelineName, strings.ToLower(sName), scheduledTime.Unix())

	// Create Job spec.
	// TODO: Stages should be able to provde most of these settings, including labels and/or annotations.
	job := &kbatch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        name,
			Namespace:   namespace,
		},
		Spec: kbatch.JobSpec{
			BackoffLimit: new(int32),
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  peanutStage.Name,
							Image: peanutStage.Spec.Image,
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}
	job.Annotations[scheduledTimeAnnotation] = scheduledTime.Format(time.RFC3339)
	job.Annotations[stageNameAnnotation] = pipelineStageName

	return job
}

func isJobFinished(job *kbatch.Job) (bool, kbatch.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if (c.Type == kbatch.JobComplete || c.Type == kbatch.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}

	return false, ""
}
