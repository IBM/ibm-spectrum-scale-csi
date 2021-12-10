---
name: Release Checklist
about: Create issue for release Checklist
title: ''
labels: ''
assignees: ''

---


## IBM Spectrum Scale CSI Driver GA release Checklist

- []  CSI snap works on Vanila k8s, Openshift with and without OLM deployment 
- []  Verify the GitHub links are working including README links 
- []  Verify KC links are working 
- []  Pause dev checkin once stop-ship phase starts
- []  Create Golden Master images on quay (Atleast 1 week before GA) 
- []  Regression Complete 
- []  Real World test complete 
- []  OLM Packages Ready and Tested
- []  Check images for Vulnerabilities(Driver, Operator, sidecars) 
- []  Final KC Verification (1 week before GA) 
- []  Create git tag draft with release content (1 week before GA)
- []  Merge dev into master (on GA date) 
- []  Push multi arch images to quay (on GA date)
- []  Publish git tag (on GA date) 
- []  Verify the images and yaml from release branch are correct - CLI install/Upgrade (on GA date) 
- []  Create PR for Community Operator Listing (on GA date) - Directory to update "upstream-community-operators" 
- []  Create PR for Community Operator Listing after "upstream-community-operators" PR is merged- Directory to update "community-operators" 
- []  Update driver listing https://kubernetes-csi.github.io/docs/drivers.html 
- []  Unpause dev checkin. (once OLM is verified for community operator) 
- []  Cloudpak certification, create case bundle, raise PR - update digest 
- []  Publish cloudpak certification 
- []  Image for ICR push 
