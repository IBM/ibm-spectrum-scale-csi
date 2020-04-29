#!/bin/python
import argparse
import sys
import os
import yaml
import zipfile
from shutil import copyfile


BASE_DIR="{0}/..".format(os.path.dirname(os.path.realpath(__file__)))
OLM_CATALOG="{0}/deploy/olm-catalog".format(BASE_DIR)
OPERATOR="{0}/ibm-spectrum-scale-csi-operator".format(OLM_CATALOG)
#PACKAGE_FILE="ibm-spectrum-scale-csi-operator.package.yaml"

PACKAGE_POSTFIX="package.yaml"

def main( args ):
  parser = argparse.ArgumentParser(
      description='''
A hack to package the operator for release.
''')

  parser.add_argument( '-d', '--dir', metavar='package dir', dest='packagedir',
    default=OPERATOR,
    help='''The manifest directory, contains a package.yaml and version directories.''')

  parser.add_argument( '-f', '--flat',  dest='flatten', action='store_true',
      help='''Flatten the bundle (for certified releases).''')

  parser.add_argument( '-o', '--output', metavar="ZIP Archive", dest='output',
    default=None,
    help='''The output location of this script. If directory supplied (or not provided) zip file
will be the <package-name>.zip (executing dir if not supplied). If a file with a path is supplied, 
it will be placed in that location.''')

  parser.add_argument( '--nozip',  dest='nozip', 
      action='store_true',
      help='''A flag to not zip the operator (useful for courier scans).''')
  
  args = parser.parse_args()

  # Find the packages
  packagedir = args.packagedir  
  if packagedir[0] is not '/':
    packagedir = "{0}/{1}".format( os.getcwd(), packagedir )

  
  # Find the package file for the supplied directory. 
  packagefile     = None
  packagefilename = None
  packages        = os.listdir(packagedir)
  for package in packages:
    if package.endswith( PACKAGE_POSTFIX ) :
      packagefilename = package
      packagefile = "{0}/{1}".format(packagedir, package)
      break
  

  # Open the package file manifest
  if packagefile is None:
    print("Unable to find package manifest.")
    return 1
  
  pkgobj={}
  with open(packagefile, 'r') as stream:
    try:
      pkgobj = yaml.safe_load(stream)
    except yaml.YAMLError as e:
      print(e)
      return 1

  # cache details. 
  packagename     = pkgobj.get("packageName"   , "package")
  defaultchannel  = pkgobj.get("defaultChannel", None) 
  selectedchannel = None

  # Get selected channels.
  for channel in pkgobj.get("channels", []):
    if channel.get("name", "") == defaultchannel:
      selectedchannel = channel 
      break

  # Determine the correct directory. 
  currentcsv = selectedchannel.get("currentCSV", None)
  csvdir     = None
  if currentcsv:
    for package in packages:
      if currentcsv.endswith( package ):
        csvdir = "{0}/{1}".format(packagedir, package)

  # Determine the zip file.
  zipname = "{0}.zip".format(packagename)
  if args.output is not None:
    if os.path.isdir(args.output):
      zipname = "{0}/{1}".format(zipname, args.output)
    else:
      zipname = args.output

  # zip the used files.
  if not args.nozip:
    with  zipfile.ZipFile(zipname, 'w') as packagezip:
      packagezip.write(packagefile, packagefilename)

      for config in os.listdir(csvdir):
        packagezip.write("{0}/{1}".format( csvdir, config), config)
  else:
    dirname= "{0}".format( args.output )
    os.mkdir(dirname)
    copyfile(packagefile, "{0}/{1}".format(dirname, packagefilename))

    for config in os.listdir(csvdir):
      copyfile("{0}/{1}".format( csvdir, config), "{0}/{1}".format(dirname, config))

  print("Package {0} was bundled to {1}".format(packagename, zipname))

if __name__ == "__main__":
  sys.exit(main(sys.argv))

