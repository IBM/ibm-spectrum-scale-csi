---
name: Bug report
about: Create a report to help us improve
title: ''
labels: 'Type: Bug'
assignees: ''

---

## Describe the bug
A clear and concise description of what the bug is.

## How to Reproduce?
Please list the steps to help development teams reproduce the behavior

1. ...


## Expected behavior
A clear and concise description of what you expected to happen.

### Data Collection and Debugging

Environmental output

- What openshift/kubernetes version are you running, and the architecture? 
- `kubectl get pods -o wide -n < csi driver namespace> `
- `kubectl get nodes -o wide`
- IBM Storage Scale container native version 
- IBM Storage Scale version 
- Output for `./tools/spectrum-scale-driver-snap.sh -n < csi driver namespace> -v `


Tool to collect the CSI snap:

`./tools/spectrum-scale-driver-snap.sh -n < csi driver namespace>`

## Screenshots
If applicable, add screenshots to help explain your problem.

## Additional context
Add any other context about the problem here.

### Add labels

- Component:
- Severity:
- Customer Impact:
- Customer Probability:
- Phase:

Note : See [labels](https://github.com/IBM/ibm-spectrum-scale-csi/labels) for the labels
