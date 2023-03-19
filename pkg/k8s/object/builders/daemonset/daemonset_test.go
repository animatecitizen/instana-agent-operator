package daemonset

import (
	"testing"

	"github.com/golang/mock/gomock"
	instanav1 "github.com/instana/instana-agent-operator/api/v1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDaemonSetBuilder_getSelectorMatchLabels(t *testing.T) {
	const name = "soijdfoijsfdoij"

	ctrl := gomock.NewController(t)

	expected := map[string]string{
		"adsf":      "eroinsvd",
		"osdgoiego": "rwuriunsv",
		"e8uriunv":  "rrudsiu",
	}

	transform := NewMockTransformations(ctrl)
	transform.EXPECT().AddCommonLabelsToMap(gomock.Eq(map[string]string{}), gomock.Eq(name), gomock.Eq(true)).Return(expected)

	d := &daemonSetBuilder{
		InstanaAgent: &instanav1.InstanaAgent{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
		Transformations: transform,
	}

	actual := d.getSelectorMatchLabels()

	assertions := require.New(t)

	assertions.Equal(expected, actual)

}

func TestDaemonSetBuilder_getPodTemplateLabels(t *testing.T) {
	t.Run("agent_mode_unset", func(t *testing.T) {
		const name = "soijdfoijsfdoij"

		ctrl := gomock.NewController(t)

		expected := map[string]string{
			"adsf":      "eroinsvd",
			"osdgoiego": "rwuriunsv",
			"e8uriunv":  "rrudsiu",
		}

		transform := NewMockTransformations(ctrl)
		transform.EXPECT().AddCommonLabelsToMap(
			gomock.Eq(map[string]string{
				"instana/agent-mode": string(instanav1.APM),
			}),
			gomock.Eq(name),
			gomock.Eq(false),
		).Return(expected)

		d := &daemonSetBuilder{
			InstanaAgent: &instanav1.InstanaAgent{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			},
			Transformations: transform,
		}

		actual := d.getPodTemplateLabels()

		assertions := require.New(t)

		assertions.Equal(expected, actual)
	})
	t.Run("agent_mode_set_by_user", func(t *testing.T) {
		const name = "soijdfoijsfdoij"

		ctrl := gomock.NewController(t)

		expected := map[string]string{
			"adsf":      "eroinsvd",
			"osdgoiego": "rwuriunsv",
			"e8uriunv":  "rrudsiu",
		}

		transform := NewMockTransformations(ctrl)
		transform.EXPECT().AddCommonLabelsToMap(
			gomock.Eq(map[string]string{
				"instana/agent-mode": string(instanav1.KUBERNETES),
			}),
			gomock.Eq(name),
			gomock.Eq(false),
		).Return(expected)

		d := &daemonSetBuilder{
			InstanaAgent: &instanav1.InstanaAgent{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: instanav1.InstanaAgentSpec{
					Agent: instanav1.BaseAgentSpec{
						Mode: instanav1.KUBERNETES,
					},
				},
			},
			Transformations: transform,
		}

		actual := d.getPodTemplateLabels()

		assertions := require.New(t)

		assertions.Equal(expected, actual)
	})
	t.Run("agent_mode_unset_with_user_given_pod_labels", func(t *testing.T) {
		const name = "soijdfoijsfdoij"

		ctrl := gomock.NewController(t)

		expected := map[string]string{
			"adsf":      "eroinsvd",
			"osdgoiego": "rwuriunsv",
			"e8uriunv":  "rrudsiu",
		}

		transform := NewMockTransformations(ctrl)
		transform.EXPECT().AddCommonLabelsToMap(
			gomock.Eq(map[string]string{
				"asdfasdf":           "eoisdgoinv",
				"reoirionv":          "98458hgoisjdf",
				"instana/agent-mode": string(instanav1.APM),
			}),
			gomock.Eq(name),
			gomock.Eq(false),
		).Return(expected)

		d := &daemonSetBuilder{
			InstanaAgent: &instanav1.InstanaAgent{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: instanav1.InstanaAgentSpec{
					Agent: instanav1.BaseAgentSpec{
						Pod: instanav1.AgentPodSpec{
							Labels: map[string]string{
								"asdfasdf":  "eoisdgoinv",
								"reoirionv": "98458hgoisjdf",
							},
						},
					},
				},
			},
			Transformations: transform,
		}

		actual := d.getPodTemplateLabels()

		assertions := require.New(t)

		assertions.Equal(expected, actual)
	})
	t.Run("agent_mode_set_by_user_with_user_given_pod_labels", func(t *testing.T) {
		const name = "soijdfoijsfdoij"

		ctrl := gomock.NewController(t)

		expected := map[string]string{
			"adsf":      "eroinsvd",
			"osdgoiego": "rwuriunsv",
			"e8uriunv":  "rrudsiu",
		}

		transform := NewMockTransformations(ctrl)
		transform.EXPECT().AddCommonLabelsToMap(
			gomock.Eq(map[string]string{
				"asdfasdf":           "eoisdgoinv",
				"reoirionv":          "98458hgoisjdf",
				"instana/agent-mode": string(instanav1.KUBERNETES),
			}),
			gomock.Eq(name),
			gomock.Eq(false),
		).Return(expected)

		d := &daemonSetBuilder{
			InstanaAgent: &instanav1.InstanaAgent{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: instanav1.InstanaAgentSpec{
					Agent: instanav1.BaseAgentSpec{
						Mode: instanav1.KUBERNETES,
						Pod: instanav1.AgentPodSpec{
							Labels: map[string]string{
								"asdfasdf":  "eoisdgoinv",
								"reoirionv": "98458hgoisjdf",
							},
						},
					},
				},
			},
			Transformations: transform,
		}

		actual := d.getPodTemplateLabels()

		assertions := require.New(t)

		assertions.Equal(expected, actual)
	})

}

func TestDaemonSetBuilder_getPodTemplateAnnotations(t *testing.T) {
	t.Run("no_user_provided_annotations", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		assertions := require.New(t)

		agent := instanav1.InstanaAgent{
			Spec: instanav1.InstanaAgentSpec{
				Cluster: instanav1.Name{
					Name: "oawgeoieg",
				},
			},
		}

		const expectedHash = "49845soidghoijw09"

		hasher := NewMockHasher(ctrl)
		hasher.EXPECT().HashOrDie(gomock.Eq(&agent.Spec)).Return(expectedHash)

		db := &daemonSetBuilder{
			InstanaAgent: &agent,
			Hasher:       hasher,
		}

		actual := db.getPodTemplateAnnotations()
		assertions.Equal(map[string]string{
			"instana-configuration-hash": expectedHash,
		}, actual)
	})
	t.Run("with_user_provided_annotations", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		assertions := require.New(t)

		agent := instanav1.InstanaAgent{
			Spec: instanav1.InstanaAgentSpec{
				Cluster: instanav1.Name{
					Name: "oawgeoieg",
				},
				Agent: instanav1.BaseAgentSpec{
					Pod: instanav1.AgentPodSpec{
						Annotations: map[string]string{
							"498hroihsg":             "4589fdoighjsoijs",
							"flkje489h309sd":         "oie409ojifg",
							"4509ufdoigjselkjweoihg": "g059pojw9jwpoijd",
						},
					},
				},
			},
		}

		const expectedHash = "49845soidghoijw09"

		hasher := NewMockHasher(ctrl)
		hasher.EXPECT().HashOrDie(gomock.Eq(&agent.Spec)).Return(expectedHash)

		db := &daemonSetBuilder{
			InstanaAgent: &agent,
			Hasher:       hasher,
		}

		actual := db.getPodTemplateAnnotations()
		assertions.Equal(map[string]string{
			"instana-configuration-hash": expectedHash,
			"498hroihsg":                 "4589fdoighjsoijs",
			"flkje489h309sd":             "oie409ojifg",
			"4509ufdoigjselkjweoihg":     "g059pojw9jwpoijd",
		}, actual)
	})
}
