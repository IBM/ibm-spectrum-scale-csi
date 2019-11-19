#!/usr/bin/env python

import sys
import os
from ConfigParser import ConfigParser
from string import Template

def configureDriver(conf_dict, infile, outfile):
        """Configures Spectrum Scale driver from config file.

        Args:
            conf_dict (dict): Dictionary of the configuration file section
            infile (string): file path of the input template file
            outfile (string): file path of the output configured file

        """

        f = open(infile, "r")
        contents = f.read()
        f.close()

        s = Template(contents)
        out = s.safe_substitute(conf_dict)

        f = open(outfile,"w+")
        f.write(out)
        f.close()

def validateCmapConfig(conf_dict):
        if conf_dict.get("clusterid") == "" or conf_dict.get("clusterid") == None :
             print "Mandatory parameter 'clusterid' in CONFIGMAP section missing"
             exit(1)

        if conf_dict.get("primaryfs") == "" or conf_dict.get("primaryfs") == None :
             print "Mandatory parameter 'primaryfs' in CONFIGMAP section missing"
             exit(1)

        if conf_dict.get("primaryfset") == "" or conf_dict.get("primaryfset") == None :
             print "Mandatory parameter 'primaryfset' in CONFIGMAP section missing"
             exit(1)

        if conf_dict.get("securesslmode") == "" or conf_dict.get("securesslmode") == None :
             print "Mandatory parameter 'securesslmode' in CONFIGMAP section missing"
             exit(1)

        if conf_dict.get("guihost") == "" or conf_dict.get("guihost") == None :
             print "Mandatory parameter 'guihost' in CONFIGMAP section missing"
             exit(1)

        if conf_dict.get("guiport") == "" or conf_dict.get("guiport") == None :
             print "Mandatory parameter 'guiport' in CONFIGMAP section missing"
             exit(1)

        if conf_dict.get("securesslmode") == "true" \
           and (conf_dict.get("cacert") == "" \
           or conf_dict.get("cacert") == None):
             print "securesslmode is true but cacert not defined"
	     exit(1)

def validateSecret(conf_dict):
        if conf_dict.get("username") == "" or conf_dict.get("username") == None \
           or conf_dict.get("password") == "" or conf_dict.get("password") == None :
             print "Mandatory base64 credentials in SECRET section missing"
             exit(1)

def validatePluginConf(conf_dict):
        if conf_dict.get("scalehostpath") == "" or conf_dict.get("scalehostpath") == None:
             print "Mandatory parameter 'scalehostpath' in PLUGIN section missing"
             exit(1)

        if conf_dict.get("namespace") == "" or conf_dict.get("namespace") == None:
             print "Mandatory parameter 'namespace' in PLUGIN section missing"
             exit(1)

def validateImages(conf_dict):
        if conf_dict.get("provisioner") == "" or conf_dict.get("provisioner") == None:
             print "Mandatory parameter 'provisioner' in IMAGES section missing"
             exit(1)

        if conf_dict.get("attacher") == "" or conf_dict.get("attacher") == None:
             print "Mandatory parameter 'attacher' in IMAGES section missing"
             exit(1)

        if conf_dict.get("driverregistrar") == "" or conf_dict.get("driverregistrar") == None:
             print "Mandatory parameter 'driverregistrar' in IMAGES section missing"
             exit(1)

        if conf_dict.get("spectrumscaleplugin") == "" or conf_dict.get("spectrumscaleplugin") == None:
             print "Mandatory parameter 'spectrumscaleplugin' in IMAGES section missing"
             exit(1)


def configureCmapConfig(conf_dict, infile, outfile):
        if conf_dict.get("inodelimit") != "" and conf_dict.get("inodelimit") != None :
             inodelimitstring = '"inodeLimit":"' + conf_dict.get("inodelimit") + '",'
             conf_dict["inodelimitstr"] = inodelimitstring
        else:
             conf_dict["inodelimitstr"] = ""

        if conf_dict.get("securesslmode") == "true":
             conf_dict["cacertstr"] = '"cacert":"guicertificate",'
        else:
             conf_dict["cacertstr"] = ""

        configureDriver(conf_dict, infile, outfile)

def configurePluginConf(conf_dict, infile, outfile, usecacert):
        if usecacert:
             conf_dict["cacertline1"] = '- name: guicertificate'
             conf_dict["cacertline2"] = 'mountPath: /var/lib/ibm/ssl/public'
             conf_dict["volcertline1"] = '- name: guicertificate'
             conf_dict["volcertline2"] = 'configMap:'
             conf_dict["volcertline3"] = 'name: guicertificate'
        else:
             conf_dict["cacertline1"] = ''
             conf_dict["cacertline2"] = ''
             conf_dict["volcertline1"] = ''
             conf_dict["volcertline2"] = ''
             conf_dict["volcertline3"] = ''

        configureDriver(conf_dict, infile, outfile)

def configure(config, section, infile, outfile):
      conf_dict = dict(config.items(section))

      if section == "CONFIGMAP":
           configureCmapConfig(conf_dict, infile, outfile)
      elif section == "SECRET":
           configureDriver(conf_dict, infile, outfile)
      elif section == "PLUGIN":
           usecacert = False
           if dict(config.items("CONFIGMAP")).get("securesslmode") == "true":
                usecacert = True

           images_dict = dict(config.items("IMAGES"))
           conf_dict["driverregistrar"] = images_dict.get("driverregistrar")
           conf_dict["spectrumscaleplugin"] = images_dict.get("spectrumscaleplugin")
           configurePluginConf(conf_dict, infile, outfile, usecacert)
      else:
           configureDriver(conf_dict, infile, outfile)

def validate(config, section):
      conf_dict = dict(config.items(section))

      if section == "CONFIGMAP":
           validateCmapConfig(conf_dict)
      elif section == "SECRET":
           validateSecret(conf_dict)
      elif section == "PLUGIN":
           validatePluginConf(conf_dict)
      else:
           validateImages(conf_dict)


def generateDeployScript(cmd, config, deployscript, deploybasepath):
      conf_dict = dict(config.items("CONFIGMAP"))

      with open(deployscript, 'w+') as f:
           f.write("set -e\n\n")

           f.write(cmd + " apply -f " + os.path.join(deploybasepath, "csi-attacher-rbac.yaml"))
           f.write("\n")
           f.write(cmd + " apply -f " + os.path.join(deploybasepath, "csi-nodeplugin-rbac.yaml"))
           f.write("\n")
           f.write(cmd + " apply -f " + os.path.join(deploybasepath, "csi-provisioner-rbac.yaml"))
           f.write("\n")

           # Create an scc only in case of openshift
           if cmd == "oc":
                f.write(cmd + " apply -f " + os.path.join(deploybasepath, "csi-plugin-scc.yaml"))
                f.write("\n")
           f.write("\n")

           f.write(cmd + " apply -f " + os.path.join(deploybasepath, "spectrum-scale-secret.json"))
           f.write("\n")

           if conf_dict.get("securesslmode") == "true":
                f.write(cmd + " create configmap guicertificate --from-file=mycertificate.pem=" + conf_dict.get("cacert") + "\n")

           f.write(cmd + " create configmap spectrum-scale-config --from-file=spectrum-scale-config.json=" + os.path.join(deploybasepath, "spectrum-scale-config.json"))
           f.write("\n")
           f.write("\n")

           f.write(cmd + " apply -f " + os.path.join(deploybasepath, "csi-plugin-attacher.yaml"))
           f.write("\n")
           f.write(cmd + " apply -f " + os.path.join(deploybasepath, "csi-plugin-provisioner.yaml"))
           f.write("\n")
           f.write(cmd + " apply -f " + os.path.join(deploybasepath, "csi-plugin.yaml"))
           f.write("\n")

      f.close()

def generateDestroyScript(cmd, config, destroyscript, deploybasepath):
      conf_dict = dict(config.items("CONFIGMAP"))

      with open(destroyscript, 'w+') as f:
           f.write("set -e\n\n")

           f.write(cmd + " delete -f " + os.path.join(deploybasepath, "csi-attacher-rbac.yaml"))
           f.write("\n") 
           f.write(cmd + " delete -f " + os.path.join(deploybasepath, "csi-nodeplugin-rbac.yaml"))
           f.write("\n") 
           f.write(cmd + " delete -f " + os.path.join(deploybasepath, "csi-provisioner-rbac.yaml"))
           f.write("\n") 

           # Delete scc in case of openshift
           if cmd == "oc":
                f.write(cmd + " delete -f " + os.path.join(deploybasepath, "csi-plugin-scc.yaml"))
                f.write("\n") 
           f.write("\n") 

           f.write(cmd + " delete -f " + os.path.join(deploybasepath, "spectrum-scale-secret.json"))
           f.write("\n") 

           if conf_dict.get("securesslmode") == "true":
                f.write(cmd + " delete configmap guicertificate \n")

           f.write(cmd + " delete configmap spectrum-scale-config \n\n")
           f.write(cmd + " delete -f " + os.path.join(deploybasepath, "csi-plugin-attacher.yaml"))
           f.write("\n") 
           f.write(cmd + " delete -f " + os.path.join(deploybasepath, "csi-plugin-provisioner.yaml"))
           f.write("\n") 
           f.write(cmd + " delete -f " + os.path.join(deploybasepath, "csi-plugin.yaml"))
           f.write("\n") 

      f.close()


if len(sys.argv) != 2 or os.path.isfile(sys.argv[1]) == False :
      print "Usage: spectrum-scale-driver.py <path_to_spectrum-scale-driver.conf>"
      print "    This command configures the Spectrum Scale Driver. Ensure that the environment variable 'CSI_SCALE_PATH' is set to the repo base path. Configure the CSI driver parameters in spectrum-scale-driver.conf and run 'spectrum-scale-driver.py <path_to_spectrum-scale-driver.conf'"
      exit(1)

basepath = ""
try:
      basepath = os.environ['CSI_SCALE_PATH']
except:
      print "CSI_SCALE_PATH not defined"
      exit(1)

deploybasepath = os.path.join(basepath, "deploy")

driverconf = sys.argv[1]
config = ConfigParser()
config.read(driverconf)

validate(config, "CONFIGMAP")
validate(config, "SECRET")
validate(config, "PLUGIN")
validate(config, "IMAGES")

configure(config, "CONFIGMAP", os.path.join(deploybasepath, "spectrum-scale-config.json_template"),
                              os.path.join(deploybasepath, "spectrum-scale-config.json"))
print "Configured '" + deploybasepath + "/spectrum-scale-config.json'"

configure(config, "SECRET", os.path.join(deploybasepath, "spectrum-scale-secret.json_template"),
                              os.path.join(deploybasepath, "spectrum-scale-secret.json"))
print "Configured '" + deploybasepath + "/spectrum-scale-secret.json'"

cmd = "kubectl"
conf_dict = dict(config.items("PLUGIN"))
if conf_dict.get("openshiftdeployment") == "true" :
    cmd = "oc"

    configure(config, "PLUGIN", os.path.join(deploybasepath, "csi-plugin-openshift.yaml_template"),
                                  os.path.join(deploybasepath, "csi-plugin.yaml"))
    print "Configured '" + deploybasepath + "/csi-plugin.yaml'"

    configure(config, "IMAGES", os.path.join(deploybasepath, "csi-plugin-provisioner-openshift.yaml_template"),
                                  os.path.join(deploybasepath, "csi-plugin-provisioner.yaml"))
    print "Configured '" + deploybasepath + "/csi-plugin-provisioner.yaml'"

    configure(config, "IMAGES", os.path.join(deploybasepath, "csi-plugin-attacher-openshift.yaml_template"),
                                  os.path.join(deploybasepath, "csi-plugin-attacher.yaml"))
    print "Configured '" + deploybasepath + "/csi-plugin-attacher.yaml'"

    configureDriver(conf_dict, os.path.join(deploybasepath, "csi-plugin-scc.yaml_template"),
                      os.path.join(deploybasepath, "csi-plugin-scc.yaml"))
else:
    cmd = "kubectl"

    configure(config, "PLUGIN", os.path.join(deploybasepath, "csi-plugin.yaml_template"),
                                  os.path.join(deploybasepath, "csi-plugin.yaml"))
    print "Configured '" + deploybasepath + "/csi-plugin.yaml'"

    configure(config, "IMAGES", os.path.join(deploybasepath, "csi-plugin-provisioner.yaml_template"),
                                  os.path.join(deploybasepath, "csi-plugin-provisioner.yaml"))
    print "Configured '" + deploybasepath + "/csi-plugin-provisioner.yaml'"

    configure(config, "IMAGES", os.path.join(deploybasepath, "csi-plugin-attacher.yaml_template"),
                                  os.path.join(deploybasepath, "csi-plugin-attacher.yaml"))
    print "Configured '" + deploybasepath + "/csi-plugin-attacher.yaml'"

configure(config, "PLUGIN", os.path.join(deploybasepath, "csi-attacher-rbac.yaml_template"),
                              os.path.join(deploybasepath, "csi-attacher-rbac.yaml"))
print "Configured '" + deploybasepath + "/csi-attacher-rbac.yaml'"

configure(config, "PLUGIN", os.path.join(deploybasepath, "csi-provisioner-rbac.yaml_template"),
                              os.path.join(deploybasepath, "csi-provisioner-rbac.yaml"))
print "Configured '" + deploybasepath + "/csi-provisioner-rbac.yaml'"

configure(config, "PLUGIN", os.path.join(deploybasepath, "csi-nodeplugin-rbac.yaml_template"),
                              os.path.join(deploybasepath, "csi-nodeplugin-rbac.yaml"))
print "Configured '" + deploybasepath + "/csi-nodeplugin-rbac.yaml'"

generateDeployScript(cmd, config, os.path.join(deploybasepath, "create.sh"), deploybasepath)
print "Generated deployment script '" + deploybasepath + "/create.sh'"

generateDestroyScript(cmd, config, os.path.join(deploybasepath, "destroy.sh"), deploybasepath)
print "Generated cleanup script '" + deploybasepath + "/destroy.sh'"

print "Spectrum Scale CSI driver configuration is complete. Please review the configuration and run '" + deploybasepath + "/create.sh' to deploy the driver"

exit(0)
