---
name: Release Checklist
about: Create issue for release Checklist
title: ''
labels: 'Type: Release Checklist'
assignees: ''

---


## IBM Storage Scale CSI Driver GA release Checklist

- [ ]  CSI snap works on Vanila k8s, Openshift with and without OLM deployment 
- [ ]  Verify GitHub links are working
- [ ]  Verify IBM docs links are working 
- [ ]  Pause dev checkin once stop-ship phase starts
- [ ]  Create Golden Master images on quay (Atleast 1 week before GA) 
- [ ]  Regression Complete 
- [ ]  Realworld test complete 
- [ ]  OLM Packages Ready and Tested
- [ ]  Create case bundle for Cloudpak certification 
- [ ]  Check images for Vulnerabilities(Driver, Operator, sidecars) 
- [ ]  IBM Docs Content Verification (1 week before GA) 
- [ ]  Create git release tag (draft) with release content (1 week before GA)
- [ ]  Merge dev into master (on GA date) 
- [ ]  Push multi arch images to quay (on GA date)
- [ ]  Publish git tag (on GA date) 
- [ ]  Verify the images and yaml from release branch are correct - CLI install/Upgrade (on GA date) 
- [ ]  Create PR for Redhat Community Operator Listing (https://github.com/redhat-openshift-ecosystem/community-operators-prod) 
- [ ]  Create PR for Community Operator Listing (https://github.com/k8s-operatorhub/community-operators)
- [ ]  Publish PR for Redhat Community Operator Listing (https://github.com/redhat-openshift-ecosystem/community-operators-prod) - (on GA date)
- [ ]  Publish PR for Community Operator Listing (https://github.com/k8s-operatorhub/community-operators) - (on GA date)
- [ ]  Update driver listing https://kubernetes-csi.github.io/docs/drivers.html
- [ ]  Unpause dev checkin. (once OLM is verified for community operator) 
- [ ]  Cloudpak certification (Content promotion ticket, Release PR, verifiy image digest)
- [ ]  Publish cloudpak certification 
- [ ]  Image for ICR push 
