FROM python:3.9-alpine
RUN apk --no-cache add curl \
        bash \
        git \
        vim

RUN git clone -b dev https://github.com/IBM/ibm-spectrum-scale-csi.git
WORKDIR "/ibm-spectrum-scale-csi/tests/functional-tests"
RUN python3.9 -m pip install -r requirements.txt
