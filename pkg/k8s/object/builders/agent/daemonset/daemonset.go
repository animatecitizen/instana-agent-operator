package daemonset

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	instanav1 "github.com/instana/instana-agent-operator/api/v1"
	"github.com/instana/instana-agent-operator/pkg/hash"
	"github.com/instana/instana-agent-operator/pkg/k8s/object/builders/common/builder"
	"github.com/instana/instana-agent-operator/pkg/k8s/object/builders/common/constants"
	"github.com/instana/instana-agent-operator/pkg/k8s/object/builders/common/env"
	"github.com/instana/instana-agent-operator/pkg/k8s/object/builders/common/helpers"
	"github.com/instana/instana-agent-operator/pkg/k8s/object/builders/common/ports"
	"github.com/instana/instana-agent-operator/pkg/k8s/object/builders/common/volume"
	"github.com/instana/instana-agent-operator/pkg/k8s/object/transformations"
	"github.com/instana/instana-agent-operator/pkg/optional"
	"github.com/instana/instana-agent-operator/pkg/pointer"
)

// TODO: Test and finish

const (
	componentName = constants.ComponentInstanaAgent
)

type daemonSetBuilder struct {
	*instanav1.InstanaAgent

	transformations.PodSelectorLabelGenerator
	hash.JsonHasher
	helpers.Helpers
	ports.PortsBuilder
	env.EnvBuilder
	volume.VolumeBuilder
}

func (d *daemonSetBuilder) ComponentName() string {
	return componentName
}

func (d *daemonSetBuilder) IsNamespaced() bool {
	return true
}

func (d *daemonSetBuilder) getPodTemplateLabels() map[string]string {
	podLabels := optional.Of(d.InstanaAgent.Spec.Agent.Pod.Labels).GetOrDefault(map[string]string{})
	podLabels[constants.LabelAgentMode] = string(optional.Of(d.InstanaAgent.Spec.Agent.Mode).GetOrDefault(instanav1.APM))

	return d.GetPodLabels(podLabels)
}

func (d *daemonSetBuilder) getEnvVars() []corev1.EnvVar {
	return d.EnvBuilder.Build(
		env.AgentModeEnv,
		env.ZoneNameEnv,
		env.ClusterNameEnv,
		env.AgentEndpointEnv,
		env.AgentEndpointPortEnv,
		env.MavenRepoURLEnv,
		env.ProxyHostEnv,
		env.ProxyPortEnv,
		env.ProxyProtocolEnv,
		env.ProxyUserEnv,
		env.ProxyPasswordEnv,
		env.ProxyUseDNSEnv,
		env.ListenAddressEnv,
		env.RedactK8sSecretsEnv,
		env.AgentKeyEnv,
		env.DownloadKeyEnv,
		env.PodNameEnv,
		env.PodIPEnv,
		env.K8sServiceDomainEnv,
	)
}

func (d *daemonSetBuilder) getContainerPorts() []corev1.ContainerPort {
	return d.GetContainerPorts(
		ports.AgentAPIsPort,
		ports.AgentSocketPort,
		ports.OpenTelemetryLegacyPort,
		ports.OpenTelemetryGRPCPort,
		ports.OpenTelemetryHTTPPort,
	)
}

func (d *daemonSetBuilder) getInitContainerVolumeMounts() []corev1.VolumeMount {
	_, res := d.VolumeBuilder.Build(volume.TPLFilesTmpVolume)
	return res
}

func (d *daemonSetBuilder) getVolumes() ([]corev1.Volume, []corev1.VolumeMount) {
	return d.VolumeBuilder.Build(
		volume.DevVolume,
		volume.RunVolume,
		volume.VarRunVolume,
		volume.VarRunKuboVolume,
		volume.VarRunContainerdVolume,
		volume.VarContainerdConfigVolume,
		volume.SysVolume,
		volume.VarLogVolume,
		volume.VarLibVolume,
		volume.VarDataVolume,
		volume.MachineIdVolume,
		volume.ConfigVolume,
		volume.TPLFilesTmpVolume,
		volume.TlsVolume,
		volume.RepoVolume,
	)
}

func (d *daemonSetBuilder) build() *appsv1.DaemonSet {
	volumes, volumeMounts := d.getVolumes()

	return &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.Name,
			Namespace: d.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: d.GetPodSelectorLabels(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      d.getPodTemplateLabels(),
					Annotations: d.InstanaAgent.Spec.Agent.Pod.Annotations,
				},
				Spec: corev1.PodSpec{
					Volumes:            volumes,
					ServiceAccountName: d.ServiceAccountName(),
					NodeSelector:       d.Spec.Agent.Pod.NodeSelector,
					HostNetwork:        true, // TODO: Test for ServiceEntry later: may not be needed on 1.26+ with internal traffic policy
					HostPID:            true,
					PriorityClassName:  d.Spec.Agent.Pod.PriorityClassName,
					DNSPolicy:          corev1.DNSClusterFirstWithHostNet,
					ImagePullSecrets:   d.ImagePullSecrets(),
					InitContainers: []corev1.Container{
						{
							Name:            "copy-tpl-files",
							Image:           d.Spec.Agent.Image(),
							ImagePullPolicy: d.Spec.Agent.PullPolicy,
							Command:         []string{"bash"},
							Args: []string{
								"-c",
								"cp " + volume.InstanaConfigDirectory + "/*.tpl " + volume.InstanaConfigTPLFilesTmpDirectory,
							},
							VolumeMounts: d.getInitContainerVolumeMounts(),
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "instana-agent",
							Image:           d.Spec.Agent.Image(),
							ImagePullPolicy: d.Spec.Agent.PullPolicy,
							Command:         []string{"bash"},
							Args: []string{
								"-c",
								"cp " + volume.InstanaConfigTPLFilesTmpDirectory + "/*.tpl " + volume.InstanaConfigDirectory + " && /opt/instana/agent/bin/run.sh",
							},
							VolumeMounts: volumeMounts,
							Env:          d.getEnvVars(),
							SecurityContext: &corev1.SecurityContext{
								Privileged: pointer.To(true),
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Host: "127.0.0.1", // TODO: Because of HostNet usage supposedly, but shouldn't be needed I think
										Path: "/status",
										Port: intstr.FromString(string(ports.AgentAPIsPort)),
									},
								},
								// TODO: set this long because startupProbe wasn't available before k8s 1.16, but this should be EOL by now, so we should see if we can revise this
								InitialDelaySeconds: 300,
								TimeoutSeconds:      3,
								PeriodSeconds:       10,
								FailureThreshold:    3,
							},
							// TODO: should have readiness probe too
							Resources: d.Spec.Agent.Pod.ResourceRequirements.GetOrDefault(),
							Ports:     d.getContainerPorts(),
						},
					},
					Tolerations: d.Spec.Agent.Pod.Tolerations,
					Affinity:    &d.Spec.Agent.Pod.Affinity,
				},
			},
			UpdateStrategy: d.InstanaAgent.Spec.Agent.UpdateStrategy,
		},
	}
}

// TODO: test Build()

func (d *daemonSetBuilder) Build() optional.Optional[client.Object] {
	if (d.Spec.Agent.Key == "" && d.Spec.Agent.KeysSecret == "") || (d.Spec.Zone.Name == "" && d.Spec.Cluster.Name == "") {
		return optional.Empty[client.Object]()
	} else {
		return optional.Of[client.Object](d.build())
	}
}

func NewDaemonSetBuilder(
	agent *instanav1.InstanaAgent,
	isOpenshift bool,
) builder.ObjectBuilder {
	return &daemonSetBuilder{
		InstanaAgent: agent,

		PodSelectorLabelGenerator: transformations.PodSelectorLabels(agent, componentName),
		JsonHasher:                hash.NewJsonHasher(),
		Helpers:                   helpers.NewHelpers(agent),
		PortsBuilder:              ports.NewPortsBuilder(agent),
		EnvBuilder:                env.NewEnvBuilder(agent),
		VolumeBuilder:             volume.NewVolumeBuilder(agent, isOpenshift),
	}
}
