# SPIFFE Link

SPIFFE Link is a batteries-included version of SPIFFE Helper. It can retrieve certificates and trust bundles from a SPIFFE Workload API endpoint, 
and load them into popular databases.

# Status
This is work in progress. Welcoming new contributors!

# Usage
`$ spiffelink --config <config_file>`

`<config_file>`: file path to the configuration file.

SPIFFE Link uses Cobra for parsing config files, so it supports Yaml, HCL, and even other configuration languages. 

# Configuration
See spiffelink_example.yaml for an example. 