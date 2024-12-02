---
name: Release Checklist
about: Create issue for release Checklist
title: ''
labels: 'Type: Release Checklist'
assignees: ''

---


## IBM Storage Scale CSI Driver GA release Checklist
- [ ]  CSI snap works on Vanila k8s
- [ ]  Verify GitHub links are working
- [ ]  Verify IBM docs links are working 
- [ ]  Pause dev checkin once stop-ship phase starts
- [ ]  Create Golden Master images on quay (Atleast 1 week before GA) 
- [ ]  Regression Complete 
- [ ]  Realworld test complete 
- [ ]  Check images for Vulnerabilities(Driver, Operator, sidecars) 
- [ ]  IBM Docs Content Verification (1 week before GA) 
- [ ]  Create git release tag (draft) with release content (1 week before GA)
- [ ]  Merge dev into master (on GA date) 
- [ ]  Push multi arch images to quay (on GA date)
- [ ]  Publish git tag (on GA date) 
- [ ]  Verify the images and yaml from release branch are correct - CLI install/Upgrade (on GA date)
- [ ]  Update driver listing https://kubernetes-csi.github.io/docs/drivers.html

