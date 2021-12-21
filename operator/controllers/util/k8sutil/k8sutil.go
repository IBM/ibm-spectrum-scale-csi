package k8sutil

import corev1 "k8s.io/api/core/v1"

// EnsureHostPathVolumeSource returns VolumeSource of type hostpath
// with given path and pathType.
func EnsureHostPathVolumeSource(path, pathType string) corev1.VolumeSource {
	t := corev1.HostPathType(pathType)
	return corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: path,
			Type: &t,
		},
	}
}

// EnsureConfigMapVolumeSource returns VolumeSource of type configMap
// with given name.
func EnsureConfigMapVolumeSource(name string) corev1.VolumeSource {
	return corev1.VolumeSource{
		ConfigMap: &corev1.ConfigMapVolumeSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
		},
	}
}

// EnsureVolume return a volume with given name and volume source.
func EnsureVolume(name string, source corev1.VolumeSource) corev1.Volume {
	return corev1.Volume{
		Name:         name,
		VolumeSource: source,
	}
}

