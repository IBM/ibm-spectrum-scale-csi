FROM python:3.7.7-alpine3.10
RUN apk --no-cache add curl \
        bash \
        git \
        vim

RUN git clone -b dev https://github.com/IBM/ibm-spectrum-scale-csi.git
WORKDIR "/ibm-spectrum-scale-csi/tests/functional-tests"
RUN python3.7 -m pip install -r requirements.txt
