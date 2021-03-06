/*
Copyright 2018 The Knative Authors

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

package testing

import (
	"fmt"
	"time"

	"github.com/knative/pkg/apis"
	"github.com/knative/pkg/apis/duck"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/knative/serving/pkg/apis/autoscaling"
	autoscalingv1alpha1 "github.com/knative/serving/pkg/apis/autoscaling/v1alpha1"
	netv1alpha1 "github.com/knative/serving/pkg/apis/networking/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	confignames "github.com/knative/serving/pkg/reconciler/v1alpha1/configuration/resources/names"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// BuildOption enables further configuration of a Build.
type BuildOption func(*unstructured.Unstructured)

// WithSucceededTrue updates the status of the provided unstructured Build object with the
// expected success condition.
func WithSucceededTrue(orig *unstructured.Unstructured) {
	cp := orig.DeepCopy()
	cp.Object["status"] = map[string]interface{}{"conditions": []duckv1alpha1.Condition{{
		Type:   duckv1alpha1.ConditionSucceeded,
		Status: corev1.ConditionTrue,
	}}}
	duck.FromUnstructured(cp, orig) // prevent panic in b.DeepCopy()
}

// WithSucceededUnknown updates the status of the provided unstructured Build object with the
// expected in-flight condition.
func WithSucceededUnknown(reason, message string) BuildOption {
	return func(orig *unstructured.Unstructured) {
		cp := orig.DeepCopy()
		cp.Object["status"] = map[string]interface{}{"conditions": []duckv1alpha1.Condition{{
			Type:    duckv1alpha1.ConditionSucceeded,
			Status:  corev1.ConditionUnknown,
			Reason:  reason,
			Message: message,
		}}}
		duck.FromUnstructured(cp, orig) // prevent panic in b.DeepCopy()
	}
}

// WithSucceededFalse updates the status of the provided unstructured Build object with the
// expected failure condition.
func WithSucceededFalse(reason, message string) BuildOption {
	return func(orig *unstructured.Unstructured) {
		cp := orig.DeepCopy()
		cp.Object["status"] = map[string]interface{}{"conditions": []duckv1alpha1.Condition{{
			Type:    duckv1alpha1.ConditionSucceeded,
			Status:  corev1.ConditionFalse,
			Reason:  reason,
			Message: message,
		}}}
		duck.FromUnstructured(cp, orig) // prevent panic in b.DeepCopy()
	}
}

// ServiceOption enables further configuration of a Service.
type ServiceOption func(*v1alpha1.Service)

var (
	// configSpec is the spec used for the different styles of Service rollout.
	configSpec = v1alpha1.ConfigurationSpec{
		RevisionTemplate: v1alpha1.RevisionTemplateSpec{
			Spec: v1alpha1.RevisionSpec{
				Container: corev1.Container{
					Image: "busybox",
				},
				TimeoutSeconds: &metav1.Duration{Duration: 60 * time.Second},
			},
		},
	}
)

// WithRunLatestRollout configures the Service to use a "runLatest" rollout.
func WithRunLatestRollout(s *v1alpha1.Service) {
	s.Spec = v1alpha1.ServiceSpec{
		RunLatest: &v1alpha1.RunLatestType{
			Configuration: configSpec,
		},
	}
}

// WithPinnedRollout configures the Service to use a "pinned" rollout,
// which is pinned to the named revision.
// Deprecated, since PinnedType is deprecated.
func WithPinnedRollout(name string) ServiceOption {
	return func(s *v1alpha1.Service) {
		s.Spec = v1alpha1.ServiceSpec{
			Pinned: &v1alpha1.PinnedType{
				RevisionName:  name,
				Configuration: configSpec,
			},
		}
	}
}

// WithReleaseRollout configures the Service to use a "release" rollout,
// which spans the provided revisions.
func WithReleaseRollout(names ...string) ServiceOption {
	return func(s *v1alpha1.Service) {
		s.Spec = v1alpha1.ServiceSpec{
			Release: &v1alpha1.ReleaseType{
				Revisions:     names,
				Configuration: configSpec,
			},
		}
	}
}

// WithManualRollout configures the Service to use a "manual" rollout.
func WithManualRollout(s *v1alpha1.Service) {
	s.Spec = v1alpha1.ServiceSpec{
		Manual: &v1alpha1.ManualType{},
	}
}

// WithInitSvcConditions initializes the Service's conditions.
func WithInitSvcConditions(s *v1alpha1.Service) {
	s.Status.InitializeConditions()
}

// WithManualStatus configures the Service to have the appropriate
// status for a "manual" rollout type.
func WithManualStatus(s *v1alpha1.Service) {
	s.Status.SetManualStatus()
}

// WithReadyRoute reflects the Route's readiness in the Service resource.
func WithReadyRoute(s *v1alpha1.Service) {
	s.Status.PropagateRouteStatus(v1alpha1.RouteStatus{
		Conditions: []duckv1alpha1.Condition{{
			Type:   "Ready",
			Status: "True",
		}},
	})
}

// WithSvcDomainStatus propagates the domain name to the status of the Service.
func WithSvcStatusDomain(s *v1alpha1.Service) {
	n, ns := s.GetName(), s.GetNamespace()
	s.Status.Domain = fmt.Sprintf("%s.%s.example.com", n, ns)
	s.Status.DomainInternal = fmt.Sprintf("%s.%s.svc.cluster.local", n, ns)
}

// WithSvcStatusAddress updates the service's status with the address.
func WithSvcStatusAddress(s *v1alpha1.Service) {
	s.Status.Address = &duckv1alpha1.Addressable{
		Hostname: fmt.Sprintf("%s.%s.svc.cluster.local", s.Name, s.Namespace),
	}
}

// WithSvcStatusTraffic sets the Service's status traffic block to the specified traffic targets.
func WithSvcStatusTraffic(traffic ...v1alpha1.TrafficTarget) ServiceOption {
	return func(r *v1alpha1.Service) {
		r.Status.Traffic = traffic
	}
}

// WithFailedRoute reflects a Route's failure in the Service resource.
func WithFailedRoute(reason, message string) ServiceOption {
	return func(s *v1alpha1.Service) {
		s.Status.PropagateRouteStatus(v1alpha1.RouteStatus{
			Conditions: []duckv1alpha1.Condition{{
				Type:    "Ready",
				Status:  "False",
				Reason:  reason,
				Message: message,
			}},
		})
	}
}

// WithReadyConfig reflects the Configuration's readiness in the Service
// resource.  This must coincide with the setting of Latest{Created,Ready}
// to the provided revision name.
func WithReadyConfig(name string) ServiceOption {
	return func(s *v1alpha1.Service) {
		s.Status.PropagateConfigurationStatus(v1alpha1.ConfigurationStatus{
			LatestCreatedRevisionName: name,
			LatestReadyRevisionName:   name,
			Conditions: []duckv1alpha1.Condition{{
				Type:   "Ready",
				Status: "True",
			}},
		})
	}
}

// WithFailedConfig reflects the Configuration's failure in the Service
// resource.  The failing revision's name is reflected in LatestCreated.
func WithFailedConfig(name, reason, message string) ServiceOption {
	return func(s *v1alpha1.Service) {
		s.Status.PropagateConfigurationStatus(v1alpha1.ConfigurationStatus{
			LatestCreatedRevisionName: name,
			Conditions: []duckv1alpha1.Condition{{
				Type:   "Ready",
				Status: "False",
				Reason: reason,
				Message: fmt.Sprintf("Revision %q failed with message: %q.",
					name, message),
			}},
		})
	}
}

// RouteOption enables further configuration of a Route.
type RouteOption func(*v1alpha1.Route)

// WithSpecTraffic sets the Route's traffic block to the specified traffic targets.
func WithSpecTraffic(traffic ...v1alpha1.TrafficTarget) RouteOption {
	return func(r *v1alpha1.Route) {
		r.Spec.Traffic = traffic
	}
}

// WithConfigTarget sets the Route's traffic block to point at a particular Configuration.
func WithConfigTarget(config string) RouteOption {
	return WithSpecTraffic(v1alpha1.TrafficTarget{
		ConfigurationName: config,
		Percent:           100,
	})
}

// WithRevTarget sets the Route's traffic block to point at a particular Revision.
func WithRevTarget(revision string) RouteOption {
	return WithSpecTraffic(v1alpha1.TrafficTarget{
		RevisionName: revision,
		Percent:      100,
	})
}

// WithStatusTraffic sets the Route's status traffic block to the specified traffic targets.
func WithStatusTraffic(traffic ...v1alpha1.TrafficTarget) RouteOption {
	return func(r *v1alpha1.Route) {
		r.Status.Traffic = traffic
	}
}

// WithDomain sets the .Status.Domain field to the prototypical domain.
func WithDomain(r *v1alpha1.Route) {
	r.Status.Domain = fmt.Sprintf("%s.%s.example.com", r.Name, r.Namespace)
}

// WithDomainInternal sets the .Status.DomainInternal field to the prototypical internal domain.
func WithDomainInternal(r *v1alpha1.Route) {
	r.Status.DomainInternal = fmt.Sprintf("%s.%s.svc.cluster.local", r.Name, r.Namespace)
}

// WithAddress sets the .Status.Address field to the prototypical internal hostname.
func WithAddress(r *v1alpha1.Route) {
	r.Status.Address = &duckv1alpha1.Addressable{
		Hostname: fmt.Sprintf("%s.%s.svc.cluster.local", r.Name, r.Namespace),
	}
}

// WithAnotherDomain sets the .Status.Domain field to an atypical domain.
func WithAnotherDomain(r *v1alpha1.Route) {
	r.Status.Domain = fmt.Sprintf("%s.%s.another-example.com", r.Name, r.Namespace)
}

// WithInitRouteConditions initializes the Service's conditions.
func WithInitRouteConditions(rt *v1alpha1.Route) {
	rt.Status.InitializeConditions()
}

// MarkTrafficAssigned calls the method of the same name on .Status
func MarkTrafficAssigned(r *v1alpha1.Route) {
	r.Status.MarkTrafficAssigned()
}

// MarkIngressReady propagates a Ready=True ClusterIngress status to the Route.
func MarkIngressReady(r *v1alpha1.Route) {
	r.Status.PropagateClusterIngressStatus(netv1alpha1.IngressStatus{
		Conditions: []duckv1alpha1.Condition{{
			Type:   "Ready",
			Status: "True",
		}},
	})
}

// MarkMissingTrafficTarget calls the method of the same name on .Status
func MarkMissingTrafficTarget(kind, revision string) RouteOption {
	return func(r *v1alpha1.Route) {
		r.Status.MarkMissingTrafficTarget(kind, revision)
	}
}

// MarkConfigurationNotReady calls the method of the same name on .Status
func MarkConfigurationNotReady(name string) RouteOption {
	return func(r *v1alpha1.Route) {
		r.Status.MarkConfigurationNotReady(name)
	}
}

// MarkConfigurationFailed calls the method of the same name on .Status
func MarkConfigurationFailed(name string) RouteOption {
	return func(r *v1alpha1.Route) {
		r.Status.MarkConfigurationFailed(name)
	}
}

// WithRouteLabel sets the specified label on the Route.
func WithRouteLabel(key, value string) RouteOption {
	return func(r *v1alpha1.Route) {
		if r.Labels == nil {
			r.Labels = make(map[string]string)
		}
		r.Labels[key] = value
	}
}

// ConfigOption enables further configuration of a Configuration.
type ConfigOption func(*v1alpha1.Configuration)

// WithBuild adds a Build to the provided Configuration.
func WithBuild(cfg *v1alpha1.Configuration) {
	cfg.Spec.Build = &v1alpha1.RawExtension{
		Object: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "testing.build.knative.dev/v1alpha1",
				"kind":       "Build",
				"spec": map[string]interface{}{
					"steps": []interface{}{
						map[string]interface{}{
							"image": "foo",
						},
						map[string]interface{}{
							"image": "bar",
						},
					},
				},
			},
		},
	}
}

// WithConfigConcurrencyModel sets the given Configuration's concurrency model.
func WithConfigConcurrencyModel(ss v1alpha1.RevisionRequestConcurrencyModelType) ConfigOption {
	return func(cfg *v1alpha1.Configuration) {
		cfg.Spec.RevisionTemplate.Spec.ConcurrencyModel = ss
	}
}

// WithGeneration sets the generation of the Configuration.
func WithGeneration(gen int64) ConfigOption {
	return func(cfg *v1alpha1.Configuration) {
		cfg.Spec.Generation = gen
	}
}

// WithObservedGen sets the observed generation of the Configuration.
func WithObservedGen(cfg *v1alpha1.Configuration) {
	cfg.Status.ObservedGeneration = cfg.Spec.Generation
}

// WithLatestCreated initializes the .status.latestCreatedRevisionName to be the name
// of the latest revision that the Configuration would have created.
func WithLatestCreated(cfg *v1alpha1.Configuration) {
	cfg.Status.SetLatestCreatedRevisionName(confignames.Revision(cfg))
}

// WithLatestReady initializes the .status.latestReadyRevisionName to be the name
// of the latest revision that the Configuration would have created.
func WithLatestReady(cfg *v1alpha1.Configuration) {
	cfg.Status.SetLatestReadyRevisionName(confignames.Revision(cfg))
}

// MarkRevisionCreationFailed calls .Status.MarkRevisionCreationFailed.
func MarkRevisionCreationFailed(msg string) ConfigOption {
	return func(cfg *v1alpha1.Configuration) {
		cfg.Status.MarkRevisionCreationFailed(msg)
	}
}

// MarkLatestCreatedFailed calls .Status.MarkLatestCreatedFailed.
func MarkLatestCreatedFailed(msg string) ConfigOption {
	return func(cfg *v1alpha1.Configuration) {
		cfg.Status.MarkLatestCreatedFailed(cfg.Status.LatestCreatedRevisionName, msg)
	}
}

// WithConfigLabel attaches a particular label to the configuration.
func WithConfigLabel(key, value string) ConfigOption {
	return func(config *v1alpha1.Configuration) {
		if config.Labels == nil {
			config.Labels = make(map[string]string)
		}
		config.Labels[key] = value
	}
}

// RevisionOption enables further configuration of a Revision.
type RevisionOption func(*v1alpha1.Revision)

// WithInitRevConditions calls .Status.InitializeConditions() on a Revision.
func WithInitRevConditions(r *v1alpha1.Revision) {
	r.Status.InitializeConditions()
}

// WithBuildRef sets the .Spec.BuildRef on the Revision to match what we'd get
// using WithBuild(name).
func WithBuildRef(name string) RevisionOption {
	return func(rev *v1alpha1.Revision) {
		rev.Spec.BuildRef = &corev1.ObjectReference{
			APIVersion: "testing.build.knative.dev/v1alpha1",
			Kind:       "Build",
			Name:       name,
		}
	}
}

// WithRevConcurrencyModel sets the concurrency model on the Revision.
func WithRevConcurrencyModel(ss v1alpha1.RevisionRequestConcurrencyModelType) RevisionOption {
	return func(rev *v1alpha1.Revision) {
		rev.Spec.ConcurrencyModel = ss
	}
}

// WithLogURL sets the .Status.LogURL to the expected value.
func WithLogURL(r *v1alpha1.Revision) {
	r.Status.LogURL = "http://logger.io/test-uid"
}

// WithCreationTimestamp sets the Revision's timestamp to the provided time.
// TODO(mattmoor): Ideally this could be a more generic Option and use meta.Accessor,
// but unfortunately Go's type system cannot support that.
func WithCreationTimestamp(t time.Time) RevisionOption {
	return func(rev *v1alpha1.Revision) {
		rev.ObjectMeta.CreationTimestamp = metav1.Time{t}
	}
}

// WithNoBuild updates the status conditions to propagate a Build status as-if
// no BuildRef was specified.
func WithNoBuild(r *v1alpha1.Revision) {
	r.Status.PropagateBuildStatus(duckv1alpha1.KResourceStatus{
		Conditions: []duckv1alpha1.Condition{{
			Type:   duckv1alpha1.ConditionSucceeded,
			Status: corev1.ConditionTrue,
			Reason: "NoBuild",
		}},
	})
}

// WithOngoingBuild propagates the status of an in-progress Build to the Revision's status.
func WithOngoingBuild(r *v1alpha1.Revision) {
	r.Status.PropagateBuildStatus(duckv1alpha1.KResourceStatus{
		Conditions: []duckv1alpha1.Condition{{
			Type:   duckv1alpha1.ConditionSucceeded,
			Status: corev1.ConditionUnknown,
		}},
	})
}

// WithSuccessfulBuild propagates the status of a successful Build to the Revision's status.
func WithSuccessfulBuild(r *v1alpha1.Revision) {
	r.Status.PropagateBuildStatus(duckv1alpha1.KResourceStatus{
		Conditions: []duckv1alpha1.Condition{{
			Type:   duckv1alpha1.ConditionSucceeded,
			Status: corev1.ConditionTrue,
		}},
	})
}

// WithFailedBuild propagates the status of a failed Build to the Revision's status.
func WithFailedBuild(reason, message string) RevisionOption {
	return func(r *v1alpha1.Revision) {
		r.Status.PropagateBuildStatus(duckv1alpha1.KResourceStatus{
			Conditions: []duckv1alpha1.Condition{{
				Type:    duckv1alpha1.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  reason,
				Message: message,
			}},
		})
	}
}

// WithEmptyLTTs clears the LastTransitionTime fields on all of the conditions of the
// provided Revision.
func WithEmptyLTTs(r *v1alpha1.Revision) {
	conds := r.Status.Conditions
	for i, c := range conds {
		// The LTT defaults and is long enough ago that we expire waiting
		// on the Endpoints to become ready.
		c.LastTransitionTime = apis.VolatileTime{}
		conds[i] = c
	}
	r.Status.SetConditions(conds)
}

// WithLastPinned updates the "last pinned" annotation to the provided timestamp.
func WithLastPinned(t time.Time) RevisionOption {
	return func(rev *v1alpha1.Revision) {
		rev.SetLastPinned(t)
	}
}

// WithRevStatus is a generic escape hatch for creating hard-to-craft
// status orientations.
func WithRevStatus(st v1alpha1.RevisionStatus) RevisionOption {
	return func(rev *v1alpha1.Revision) {
		rev.Status = st
	}
}

// MarkActive calls .Status.MarkActive on the Revision.
func MarkActive(r *v1alpha1.Revision) {
	r.Status.MarkActive()
}

// MarkInactive calls .Status.MarkInactive on the Revision.
func MarkInactive(reason, message string) RevisionOption {
	return func(r *v1alpha1.Revision) {
		r.Status.MarkInactive(reason, message)
	}
}

// MarkActivating calls .Status.MarkActivating on the Revision.
func MarkActivating(reason, message string) RevisionOption {
	return func(r *v1alpha1.Revision) {
		r.Status.MarkActivating(reason, message)
	}
}

// MarkDeploying calls .Status.MarkDeploying on the Revision.
func MarkDeploying(reason string) RevisionOption {
	return func(r *v1alpha1.Revision) {
		r.Status.MarkDeploying(reason)
	}
}

// MarkProgressDeadlineExceeded calls the method of the same name on the Revision
// with the message we expect the Revision Reconciler to pass.
func MarkProgressDeadlineExceeded(r *v1alpha1.Revision) {
	r.Status.MarkProgressDeadlineExceeded("Unable to create pods for more than 120 seconds.")
}

// MarkServiceTimeout calls .Status.MarkServiceTimeout on the Revision.
func MarkServiceTimeout(r *v1alpha1.Revision) {
	r.Status.MarkServiceTimeout()
}

// MarkContainerMissing calls .Status.MarkContainerMissing on the Revision.
func MarkContainerMissing(rev *v1alpha1.Revision) {
	rev.Status.MarkContainerMissing("It's the end of the world as we know it")
}

// MarkRevisionReady calls the necessary helpers to make the Revision Ready=True.
func MarkRevisionReady(r *v1alpha1.Revision) {
	WithInitRevConditions(r)
	WithNoBuild(r)
	MarkActive(r)
	r.Status.MarkResourcesAvailable()
	r.Status.MarkContainerHealthy()
}

type PodAutoscalerOption func(*autoscalingv1alpha1.PodAutoscaler)

// WithTraffic updates the PA to reflect it receiving traffic.
func WithTraffic(pa *autoscalingv1alpha1.PodAutoscaler) {
	pa.Status.MarkActive()
}

// WithBufferedTraffic updates the PA to reflect that it has received
// and buffered traffic while it is being activated.
func WithBufferedTraffic(reason, message string) PodAutoscalerOption {
	return func(pa *autoscalingv1alpha1.PodAutoscaler) {
		pa.Status.MarkActivating(reason, message)
	}
}

// WithNoTraffic updates the PA to reflect the fact that it is not
// receiving traffic.
func WithNoTraffic(reason, message string) PodAutoscalerOption {
	return func(pa *autoscalingv1alpha1.PodAutoscaler) {
		pa.Status.MarkInactive(reason, message)
	}
}

// WithHPAClass updates the PA to add the hpa class annotation.
func WithHPAClass(pa *autoscalingv1alpha1.PodAutoscaler) {
	if pa.Annotations == nil {
		pa.Annotations = make(map[string]string)
	}
	pa.Annotations[autoscaling.ClassAnnotationKey] = autoscaling.HPA
}

// WithKPAClass updates the PA to add the kpa class annotation.
func WithKPAClass(pa *autoscalingv1alpha1.PodAutoscaler) {
	if pa.Annotations == nil {
		pa.Annotations = make(map[string]string)
	}
	pa.Annotations[autoscaling.ClassAnnotationKey] = autoscaling.KPA
}

// WithTargetAnnotation adds a target annotation to the PA.
func WithTargetAnnotation(pa *autoscalingv1alpha1.PodAutoscaler) {
	if pa.Annotations == nil {
		pa.Annotations = make(map[string]string)
	}
	pa.Annotations[autoscaling.TargetAnnotationKey] = "50"
}

// WithMetricAnnotation adds a metric annotation to the PA.
func WithMetricAnnotation(metric string) PodAutoscalerOption {
	return func(pa *autoscalingv1alpha1.PodAutoscaler) {
		if pa.Annotations == nil {
			pa.Annotations = make(map[string]string)
		}
		pa.Annotations[autoscaling.MetricAnnotationKey] = metric
	}
}

// K8sServiceOption enables further configuration of the Kubernetes Service.
type K8sServiceOption func(*corev1.Service)

// MutateK8sService changes the service in a way that must be reconciled.
func MutateK8sService(svc *corev1.Service) {
	// An effective hammer ;-P
	svc.Spec = corev1.ServiceSpec{}
}

func WithClusterIP(ip string) K8sServiceOption {
	return func(svc *corev1.Service) {
		svc.Spec.ClusterIP = ip
	}
}

func WithExternalName(name string) K8sServiceOption {
	return func(svc *corev1.Service) {
		svc.Spec.ExternalName = name
	}
}

// EndpointsOption enables further configuration of the Kubernetes Endpoints.
type EndpointsOption func(*corev1.Endpoints)

// WithSubsets adds subsets to the body of a Revision, enabling us to refer readiness.
func WithSubsets(ep *corev1.Endpoints) {
	ep.Subsets = []corev1.EndpointSubset{{
		Addresses: []corev1.EndpointAddress{{IP: "127.0.0.1"}},
	}}
}
