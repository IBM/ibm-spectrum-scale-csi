import os
import argparse


def upgrade(frm,to):

    if frm == "2.4.0" and to == "2.5.0":
        print(f"UPGRADING FROM {frm} TO {to} ...")

        os.system("curl -O https://raw.githubusercontent.com/IBM/ibm-spectrum-scale-csi/v2.5.0/generated/installer/ibm-spectrum-scale-csi-operator.yaml")
        os.system("kubectl apply -f ibm-spectrum-scale-csi-operator.yaml")
        os.system("kubectl get pod -n ibm-spectrum-scale-csi-driver")
        os.system("kubectl describe pod ibm-spectrum-scale-csi-operator-8947b76cb-k8ggx -n ibm-spectrum-scale-csi-driver")

    elif frm == "2.4.0" and to == "2.5.1":
        print(f"UPGRADING FROM {frm} TO {to} ...")

        os.system("curl -O https://raw.githubusercontent.com/IBM/ibm-spectrum-scale-csi/v2.5.1/generated/installer/ibm-spectrum-scale-csi-operator.yaml")
        os.system("kubectl apply -f ibm-spectrum-scale-csi-operator.yaml")
        os.system("kubectl get pod -n ibm-spectrum-scale-csi-driver")
        os.system("kubectl describe pod ibm-spectrum-scale-csi-operator-8947b76cb-k8ggx -n ibm-spectrum-scale-csi-driver")

    elif frm == None or to == None:
        print(f"YOU NEED TO USE -from=\"version\" and -to=\"version\" TO SPECIFY THE VERSIONS...")
    else:
        print(f"UPGRADING FROM {frm} TO {to} IS NOT SUPPORTED...")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="This program will do IBM upgrades.",formatter_class=argparse.ArgumentDefaultsHelpFormatter)

    parser.add_argument("-from", help = "From version (format: N.N.N)",default="2.4.0")
    parser.add_argument("-to", help = "From version (format: N.N.N)",default="2.5.1")
    args = parser.parse_args()
    config = vars(args)

    upgrade(config["from"],config["to"])
