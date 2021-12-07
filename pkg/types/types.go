package types

import (
	"errors"
	"strings"

	"github.com/grafana/k6-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Internal type created to support Spec.script options
type Script struct {
	Name string
	File string
	Type string // ConfigMap or VolumeClaim
}

func ParseScript(spec *v1alpha1.K6Spec) (*Script, error) {
	s := &Script{}
	s.File = "test.js"

	if spec.Script.VolumeClaim.Name != "" {
		s.Name = spec.Script.VolumeClaim.Name
		if spec.Script.VolumeClaim.File != "" {
			s.File = spec.Script.VolumeClaim.File
		}

		s.Type = "VolumeClaim"
		return s, nil
	}

	if spec.Script.ConfigMap.Name != "" {
		s.Name = spec.Script.ConfigMap.Name

		if spec.Script.ConfigMap.File != "" {
			s.File = spec.Script.ConfigMap.File
		}

		s.Type = "ConfigMap"
		return s, nil
	}

	return nil, errors.New("ConfigMap or VolumeClaim not provided in script definition")
}

func (s *Script) Volume() corev1.Volume {
	if s.Type == "VolumeClaim" {
		return corev1.Volume{
			Name: "k6-test-volume",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: s.Name,
				},
			},
		}
	}

	return corev1.Volume{
		Name: "k6-test-volume",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: s.Name,
				},
			},
		},
	}
}

// Internal type to support k6 invocation in initialization stage.
// Not all k6 commands allow the same set of arguments so CLI is object meant to contain only the ones fit for the arhive call.
// Maybe revise this once crococonf is closer to integration?
type CLI struct {
	ArchiveArgs string
	// k6-operator doesn't care for most values of CLI arguments to k6, with an exception of cloud output
	HasCloudOut bool
}

func ParseCLI(spec *v1alpha1.K6Spec) *CLI {
	lastArgV := func(start int, args []string) (end int) {
		var nextArg bool
		end = start + 1
		for !nextArg && end < len(args) {
			if args[end][0] == '-' {
				nextArg = true
				break
			}
			end++
		}
		return
	}

	var cli CLI

	args := strings.Split(spec.Arguments, " ")
	i := 0
	for i < len(args) {
		args[i] = strings.TrimSpace(args[i])
		if len(args[i]) == 0 {
			i++
			continue
		}
		if args[i][0] == '-' {
			end := lastArgV(i+1, args)

			switch args[i] {
			case "-o", "--out":
				for j := 0; j < end; j++ {
					if args[j] == "cloud" {
						cli.HasCloudOut = true
					}
				}
			case "-l", "--linger", "--no-usage-report":
				// non-archive arguments, so skip them
				break
			default:
				if len(cli.ArchiveArgs) > 0 {
					cli.ArchiveArgs += " "
				}
				cli.ArchiveArgs += strings.Join(args[i:end], " ")
			}
			i = end
		}
	}

	return &cli
}
