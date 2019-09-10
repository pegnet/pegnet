# Oracle Price Record (OPR) Decoding Module

This module contains all the necessary information to decode the content of the OPRs on the PegNet chains. There are two different formats, JSON and Protobuf, as well as two different lists of assets. 

## JSON

Contains a custom marshaller to ensure that the order of assets are the same as V1Assets. 

## PegNet MainNet

| Version | Asset List | OPR Format |
|---|---|---|
| 1 | V1 | JSON |
| 2 | V2 | Protobuf |