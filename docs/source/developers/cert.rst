Cert Process
============

Creating the Pull Request
-------------------------

1. Fork `https://github.ibm.com/IBMPrivateCloud/charts`_.
2. Clone the forked repository .
3. From the root dir (of this project) execute the following:

.. code-block:: bash
  
  export CHARTS=<Local Chart Repo Root>
  
  cp -R -L cloudpak/ ${CHARTS}
  cd ${CHARTS}/stable

  git checkout -b  ibm-spectrum-scale-csi-operator-bundle
  git add ibm-spectrum-scale-csi-operator-bundle
  git commit -S -m "Some message"
  git push origin ibm-spectrum-scale-csi-operator-bundle

4. Follow standard Pull Request procedures.
