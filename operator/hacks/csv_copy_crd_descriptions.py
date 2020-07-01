#!/bin/python

import argparse
import sys
import os
import yaml

BASE_DIR="{0}/../".format(os.path.dirname(os.path.realpath(__file__)))
DEFAULT_VERSION="2.0.0"
CSV_PATH="{0}deploy/olm-catalog/ibm-spectrum-scale-csi-operator/{1}/ibm-spectrum-scale-csi-operator.v{1}.clusterserviceversion.yaml"
CRD_SOURCE_PATH="{0}deploy/crds/{1}"
CRD_TARGET_PATH="{0}deploy/olm-catalog/ibm-spectrum-scale-csi-operator/{1}/{2}"

def loadDescriptors(descType, crd, csv, descriptorMap={}):
    props = crd.get("spec", {})    \
        .get("validation",{})      \
        .get("openAPIV3Schema", {})\
        .get("properties", {})     \
    
    spec            = props.get(descType, {})
    specprops       = spec.get("properties", {})
    specdescriptors = []
    specpaths       = specprops.keys()
    
    while len(specpaths) > 0:
        # Load the path and then iterate to get the property.
        path    = specpaths.pop(0)
        keys    = path.split(".")
        name    = keys[-1]
        prop    = spec
        subprops = specprops
        displayName = name
  
        for key in keys:
            prop     = subprops.get(key, {})
            subprops = prop.get("items",{}).get("properties", {})

        
        # Enqueue the sub properties (if present)
        for subprop in subprops.keys():
            specpaths.append("{0}.{1}".format(path,subprop))
        
        
        desc = descriptorMap.get(path,{})
        # Construct description
        specdescriptors.append({ 
          "displayName" : desc.get("displayName", displayName),
          "x-descriptors": desc.get("x-descriptors", []),
          "path" : path,
          "description" : prop.get("description", "") })
        
    return specdescriptors

def mapDescriptors(metaname, spec):
  specMap = {}
  statusMap = {}
  for resource in spec:
    if resource.get("name","") == metaname:
      for specD in resource.get("specDescriptors",[]):
        specMap[specD.get("path","")] = specD

      for statusD in resource.get("statusDescriptors",[]):
        statusMap[statusD.get("path","")] = statusD

  return (specMap, statusMap)

def main(args):
    parser = argparse.ArgumentParser(
        description='''A hack to clone descriptions from the CRD.''')
    
    parser.add_argument( '--crd', metavar='crd', dest='crd', default=None,
        help='''The Custom Resource Definition File.''')

    parser.add_argument( '--version', metavar='CSV Version', dest='version', default=DEFAULT_VERSION,
      help='''The version of the CSV to update''')

    args = parser.parse_args()

    
    crdf = CRD_SOURCE_PATH.format(BASE_DIR, args.crd)
    csvf = CSV_PATH.format(BASE_DIR, args.version)

    crd = None
    csv = None
    try:
        with open(crdf, 'r') as stream:
            crd = yaml.safe_load(stream)
        with open(csvf, 'r') as stream:
            csv = yaml.safe_load(stream)
    except yaml.YAMLError as e:
        print(e)
        return 1

    if crd is not None and csv is not None:
        metaname= crd.get("metadata",{}).get("name", "")

        # TODO need to replace with something 
        resourcefound=False
        owned = csv.get("spec",{}).get("customresourcedefinitions",{}).get("owned",{})
        specmap, statusmap =  mapDescriptors(metaname, owned)

        specdescriptors = loadDescriptors("spec", crd, csv, specmap)
        statusdescriptors = loadDescriptors("status", crd, csv, statusmap)

        for resource in owned:
            if resource.get("name","") == metaname:
                resourcefound = True
                resource["specDescriptors"] = specdescriptors
                resource["statusDescriptors"] = statusdescriptors
                resource["version"] = crd.get("spec",{}).get("version","v1")

        # If the resource wasn't present in the customresourcedefinitions add it. 
        if not resourcefound:
            owned.append({
                "name"            : metaname,
                "kind"            : crd.get("kind","CustomResourceDefinition"),
                "version"         : crd.get("spec",{}).get("version","v1"),
                "displayName"     : metaname,
                "specDescriptors" : specdescriptors,
                "description"     : "TODO: Fill this in"})

        # Copy the updated CSV
        with open(csvf, 'w') as outfile:
            yaml.dump(csv, outfile, default_flow_style=False)

        # Copy the CRD
        crdtarget=CRD_TARGET_PATH.format(BASE_DIR, args.version, os.path.basename(crdf))
        with open(crdtarget, 'w') as outfile:
          yaml.dump(crd, outfile, default_flow_style=False)
        
if __name__ == "__main__":
    sys.exit(main(sys.argv))

