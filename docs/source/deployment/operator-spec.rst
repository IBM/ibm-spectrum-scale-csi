Contents
--------

.. contents::
   :local:

Operator Scope
==============

The IBM Spectrum Scale CSI Operator is a cluster scoped operator at this time. Most operator operations
are limited to the deployed namespace of the operator, however, the underlying Driver requires 
cluster level role bindings.

Additionally, the operator requires cluster level access to `securitycontextconstraints` to manage
the security constraints of the operator in OpenShift deployments.


