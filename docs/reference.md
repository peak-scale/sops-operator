# Reference

Packages:

- [addons.projectcapsule.dev/v1alpha1](#addonsprojectcapsuledevv1alpha1)

# addons.projectcapsule.dev/v1alpha1

Resource Types:

- [GlobalSopsSecret](#globalsopssecret)

- [SopsProvider](#sopsprovider)

- [SopsSecret](#sopssecret)




## GlobalSopsSecret






GlobalSopsSecret is the Schema for the globalsopssecrets API.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **apiVersion** | string | addons.projectcapsule.dev/v1alpha1 | true |
| **kind** | string | GlobalSopsSecret | true |
| **[metadata](https://kubernetes.io/docs/reference/generated/kubernetes-api/latest/#objectmeta-v1-meta)** | object | Refer to the Kubernetes API documentation for the fields of the `metadata` field. | true |
| **[sops](#globalsopssecretsops)** | object | Metadata is stored in SOPS encrypted files, and it contains the information necessary to decrypt the file.
This struct is just used for serialization, and SOPS uses another struct internally, sops.Metadata. It exists
in order to allow the binary format to stay backwards compatible over time, but at the same time allow the internal
representation SOPS uses to change over time. | true |
| **[spec](#globalsopssecretspec)** | object | SopsSecretSpec defines the desired state of SopsSecret. | false |
| **[status](#globalsopssecretstatus)** | object | SopsSecretStatus defines the observed state of SopsSecret. | false |


### GlobalSopsSecret.sops



Metadata is stored in SOPS encrypted files, and it contains the information necessary to decrypt the file.
This struct is just used for serialization, and SOPS uses another struct internally, sops.Metadata. It exists
in order to allow the binary format to stay backwards compatible over time, but at the same time allow the internal
representation SOPS uses to change over time.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **lastmodified** | string |  | true |
| **mac** | string |  | true |
| **[age](#globalsopssecretsopsageindex)** | []object |  | false |
| **[azure_kv](#globalsopssecretsopsazure_kvindex)** | []object |  | false |
| **encrypted_comment_regex** | string |  | false |
| **encrypted_regex** | string |  | false |
| **encrypted_suffix** | string |  | false |
| **[gcp_kms](#globalsopssecretsopsgcp_kmsindex)** | []object |  | false |
| **[hc_vault](#globalsopssecretsopshc_vaultindex)** | []object |  | false |
| **[key_groups](#globalsopssecretsopskey_groupsindex)** | []object |  | false |
| **[kms](#globalsopssecretsopskmsindex)** | []object |  | false |
| **mac_only_encrypted** | boolean |  | false |
| **[pgp](#globalsopssecretsopspgpindex)** | []object |  | false |
| **shamir_threshold** | integer |  | false |
| **unencrypted_comment_regex** | string |  | false |
| **unencrypted_regex** | string |  | false |
| **unencrypted_suffix** | string |  | false |
| **version** | string |  | false |


### GlobalSopsSecret.sops.age[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **enc** | string |  | true |
| **recipient** | string |  | true |


### GlobalSopsSecret.sops.azure_kv[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **name** | string |  | true |
| **vault_url** | string |  | true |
| **version** | string |  | true |


### GlobalSopsSecret.sops.gcp_kms[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **resource_id** | string |  | true |


### GlobalSopsSecret.sops.hc_vault[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **engine_path** | string |  | true |
| **key_name** | string |  | true |
| **vault_address** | string |  | true |


### GlobalSopsSecret.sops.key_groups[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[age](#globalsopssecretsopskey_groupsindexageindex)** | []object |  | false |
| **[azure_kv](#globalsopssecretsopskey_groupsindexazure_kvindex)** | []object |  | false |
| **[gcp_kms](#globalsopssecretsopskey_groupsindexgcp_kmsindex)** | []object |  | false |
| **[hc_vault](#globalsopssecretsopskey_groupsindexhc_vaultindex)** | []object |  | false |
| **[kms](#globalsopssecretsopskey_groupsindexkmsindex)** | []object |  | false |
| **[pgp](#globalsopssecretsopskey_groupsindexpgpindex)** | []object |  | false |


### GlobalSopsSecret.sops.key_groups[index].age[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **enc** | string |  | true |
| **recipient** | string |  | true |


### GlobalSopsSecret.sops.key_groups[index].azure_kv[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **name** | string |  | true |
| **vault_url** | string |  | true |
| **version** | string |  | true |


### GlobalSopsSecret.sops.key_groups[index].gcp_kms[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **resource_id** | string |  | true |


### GlobalSopsSecret.sops.key_groups[index].hc_vault[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **engine_path** | string |  | true |
| **key_name** | string |  | true |
| **vault_address** | string |  | true |


### GlobalSopsSecret.sops.key_groups[index].kms[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **arn** | string |  | true |
| **aws_profile** | string |  | true |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **context** | map[string]string |  | false |
| **role** | string |  | false |


### GlobalSopsSecret.sops.key_groups[index].pgp[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | false |
| **enc** | string |  | false |
| **fp** | string |  | false |


### GlobalSopsSecret.sops.kms[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **arn** | string |  | true |
| **aws_profile** | string |  | true |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **context** | map[string]string |  | false |
| **role** | string |  | false |


### GlobalSopsSecret.sops.pgp[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | false |
| **enc** | string |  | false |
| **fp** | string |  | false |


### GlobalSopsSecret.spec



SopsSecretSpec defines the desired state of SopsSecret.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[secrets](#globalsopssecretspecsecretsindex)** | []object | Define Secrets to replicate, when secret is decrypted | true |


### GlobalSopsSecret.spec.secrets[index]



GlobalSopsSecretItem defines the desired state of GlobalSopsSecret.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **name** | string | Name must be unique within a namespace. Is required when creating resources, although
some resources may allow a client to request the generation of an appropriate name
automatically. Name is primarily intended for creation idempotence and configuration
definition.
Cannot be updated.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names | true |
| **namespace** | string | Namespace must be declared since this is a cluster scoped resource | true |
| **annotations** | map[string]string | Map of string keys and values that can be used to organize and categorize
(scope and select) objects. May match selectors of replication controllers
and services.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels | false |
| **data** | map[string]string | Data map to use in Kubernetes secret (equivalent to Kubernetes Secret object data, please see for more
information: https://kubernetes.io/docs/concepts/configuration/secret/#overview-of-secrets) | false |
| **immutable** | boolean | Immutable, if set to true, ensures that data stored in the Secret cannot
be updated (only object metadata can be modified).
If not set to true, the field can be modified at any time.
Defaulted to nil. | false |
| **labels** | map[string]string | Map of string keys and values that can be used to organize and categorize
(scope and select) objects. May match selectors of replication controllers
and services.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels | false |
| **stringData** | map[string]string | stringData map to use in Kubernetes secret (equivalent to Kubernetes Secret object stringData, please see for more
information: https://kubernetes.io/docs/concepts/configuration/secret/#overview-of-secrets) | false |
| **type** | enum | Kubernetes secret type.
Defaults to Opaque.
Allowed values:
- Opaque
- kubernetes.io/service-account-token
- kubernetes.io/dockercfg
- kubernetes.io/dockerconfigjson
- kubernetes.io/basic-auth
- kubernetes.io/ssh-auth
- kubernetes.io/tls
- bootstrap.kubernetes.io/token<br/><i>Enum</i>: Opaque, kubernetes.io/service-account-token, kubernetes.io/dockercfg, kubernetes.io/dockerconfigjson, kubernetes.io/basic-auth, kubernetes.io/ssh-auth, kubernetes.io/tls, bootstrap.kubernetes.io/token<br/> | false |


### GlobalSopsSecret.status



SopsSecretStatus defines the observed state of SopsSecret.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[condition](#globalsopssecretstatuscondition)** | object | Conditions represent the latest available observations of an instances state | false |
| **[providers](#globalsopssecretstatusprovidersindex)** | []object | Providers used on this secret | false |
| **[secrets](#globalsopssecretstatussecretsindex)** | []object | Secrets being replicated by this SopsSecret | false |
| **size** | integer | Amount of Secrets<br/><i>Default</i>: 0<br/> | false |


### GlobalSopsSecret.status.condition



Conditions represent the latest available observations of an instances state

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **lastTransitionTime** | string | lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/><i>Format</i>: date-time<br/> | true |
| **message** | string | message is a human readable message indicating details about the transition.
This may be an empty string. | true |
| **reason** | string | reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty. | true |
| **status** | enum | status of the condition, one of True, False, Unknown.<br/><i>Enum</i>: True, False, Unknown<br/> | true |
| **type** | string | type of condition in CamelCase or in foo.example.com/CamelCase. | true |
| **observedGeneration** | integer | observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/><i>Format</i>: int64<br/><i>Minimum</i>: 0<br/> | false |


### GlobalSopsSecret.status.providers[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **name** | string | Name of Object | true |
| **namespace** | string | namespace of Object | false |
| **uid** | string | namespace of Object | false |


### GlobalSopsSecret.status.secrets[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[condition](#globalsopssecretstatussecretsindexcondition)** | object | Condition contains details for one aspect of the current state of this API Resource. | true |
| **name** | string |  | true |
| **namespace** | string |  | true |
| **uid** | string | UID is a type that holds unique ID values, including UUIDs.  Because we
don't ONLY use UUIDs, this is an alias to string.  Being a type captures
intent and helps make sure that UIDs and names do not get conflated. | false |


### GlobalSopsSecret.status.secrets[index].condition



Condition contains details for one aspect of the current state of this API Resource.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **lastTransitionTime** | string | lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/><i>Format</i>: date-time<br/> | true |
| **message** | string | message is a human readable message indicating details about the transition.
This may be an empty string. | true |
| **reason** | string | reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty. | true |
| **status** | enum | status of the condition, one of True, False, Unknown.<br/><i>Enum</i>: True, False, Unknown<br/> | true |
| **type** | string | type of condition in CamelCase or in foo.example.com/CamelCase. | true |
| **observedGeneration** | integer | observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/><i>Format</i>: int64<br/><i>Minimum</i>: 0<br/> | false |

## SopsProvider






SopsProvider is the Schema for the sopsproviders API.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **apiVersion** | string | addons.projectcapsule.dev/v1alpha1 | true |
| **kind** | string | SopsProvider | true |
| **[metadata](https://kubernetes.io/docs/reference/generated/kubernetes-api/latest/#objectmeta-v1-meta)** | object | Refer to the Kubernetes API documentation for the fields of the `metadata` field. | true |
| **[spec](#sopsproviderspec)** | object | SopsProviderSpec defines the desired state of SopsProvider. | false |
| **[status](#sopsproviderstatus)** | object | SopsProviderStatus defines the observed state of SopsProvider. | false |


### SopsProvider.spec



SopsProviderSpec defines the desired state of SopsProvider.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[keys](#sopsproviderspeckeysindex)** | []object | Select namespaces or secrets where decryption information for this
provider can be sourced from | true |
| **[sops](#sopsproviderspecsopsindex)** | []object | Selector Referencing which Secrets can be encrypted by this provider
This selects effective SOPS Secrets | true |


### SopsProvider.spec.keys[index]



Selector for resources and their labels or selecting origin namespaces

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[matchExpressions](#sopsproviderspeckeysindexmatchexpressionsindex)** | []object | matchExpressions is a list of label selector requirements. The requirements are ANDed. | false |
| **matchLabels** | map[string]string | matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed. | false |
| **[namespaceSelector](#sopsproviderspeckeysindexnamespaceselector)** | object | NamespaceSelector for filtering namespaces by labels where items can be located in | false |


### SopsProvider.spec.keys[index].matchExpressions[index]



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **key** | string | key is the label key that the selector applies to. | true |
| **operator** | string | operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist. | true |
| **values** | []string | values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch. | false |


### SopsProvider.spec.keys[index].namespaceSelector



NamespaceSelector for filtering namespaces by labels where items can be located in

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[matchExpressions](#sopsproviderspeckeysindexnamespaceselectormatchexpressionsindex)** | []object | matchExpressions is a list of label selector requirements. The requirements are ANDed. | false |
| **matchLabels** | map[string]string | matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed. | false |


### SopsProvider.spec.keys[index].namespaceSelector.matchExpressions[index]



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **key** | string | key is the label key that the selector applies to. | true |
| **operator** | string | operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist. | true |
| **values** | []string | values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch. | false |


### SopsProvider.spec.sops[index]



Selector for resources and their labels or selecting origin namespaces

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[matchExpressions](#sopsproviderspecsopsindexmatchexpressionsindex)** | []object | matchExpressions is a list of label selector requirements. The requirements are ANDed. | false |
| **matchLabels** | map[string]string | matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed. | false |
| **[namespaceSelector](#sopsproviderspecsopsindexnamespaceselector)** | object | NamespaceSelector for filtering namespaces by labels where items can be located in | false |


### SopsProvider.spec.sops[index].matchExpressions[index]



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **key** | string | key is the label key that the selector applies to. | true |
| **operator** | string | operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist. | true |
| **values** | []string | values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch. | false |


### SopsProvider.spec.sops[index].namespaceSelector



NamespaceSelector for filtering namespaces by labels where items can be located in

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[matchExpressions](#sopsproviderspecsopsindexnamespaceselectormatchexpressionsindex)** | []object | matchExpressions is a list of label selector requirements. The requirements are ANDed. | false |
| **matchLabels** | map[string]string | matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed. | false |


### SopsProvider.spec.sops[index].namespaceSelector.matchExpressions[index]



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **key** | string | key is the label key that the selector applies to. | true |
| **operator** | string | operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist. | true |
| **values** | []string | values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch. | false |


### SopsProvider.status



SopsProviderStatus defines the observed state of SopsProvider.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[condition](#sopsproviderstatuscondition)** | object | Conditions represent the latest available observations of an instances state | false |
| **[providers](#sopsproviderstatusprovidersindex)** | []object | List Validated Providers | false |
| **size** | integer | Amount of providers<br/><i>Default</i>: 0<br/> | false |


### SopsProvider.status.condition



Conditions represent the latest available observations of an instances state

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **lastTransitionTime** | string | lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/><i>Format</i>: date-time<br/> | true |
| **message** | string | message is a human readable message indicating details about the transition.
This may be an empty string. | true |
| **reason** | string | reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty. | true |
| **status** | enum | status of the condition, one of True, False, Unknown.<br/><i>Enum</i>: True, False, Unknown<br/> | true |
| **type** | string | type of condition in CamelCase or in foo.example.com/CamelCase. | true |
| **observedGeneration** | integer | observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/><i>Format</i>: int64<br/><i>Minimum</i>: 0<br/> | false |


### SopsProvider.status.providers[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **name** | string | Name of Object | true |
| **[condition](#sopsproviderstatusprovidersindexcondition)** | object | Conditions represent the latest available observations of an instances state | false |
| **namespace** | string | namespace of Object | false |
| **uid** | string | namespace of Object | false |


### SopsProvider.status.providers[index].condition



Conditions represent the latest available observations of an instances state

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **lastTransitionTime** | string | lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/><i>Format</i>: date-time<br/> | true |
| **message** | string | message is a human readable message indicating details about the transition.
This may be an empty string. | true |
| **reason** | string | reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty. | true |
| **status** | enum | status of the condition, one of True, False, Unknown.<br/><i>Enum</i>: True, False, Unknown<br/> | true |
| **type** | string | type of condition in CamelCase or in foo.example.com/CamelCase. | true |
| **observedGeneration** | integer | observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/><i>Format</i>: int64<br/><i>Minimum</i>: 0<br/> | false |

## SopsSecret






SopsSecret is the Schema for the sopssecrets API.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **apiVersion** | string | addons.projectcapsule.dev/v1alpha1 | true |
| **kind** | string | SopsSecret | true |
| **[metadata](https://kubernetes.io/docs/reference/generated/kubernetes-api/latest/#objectmeta-v1-meta)** | object | Refer to the Kubernetes API documentation for the fields of the `metadata` field. | true |
| **[sops](#sopssecretsops)** | object | Metadata is stored in SOPS encrypted files, and it contains the information necessary to decrypt the file.
This struct is just used for serialization, and SOPS uses another struct internally, sops.Metadata. It exists
in order to allow the binary format to stay backwards compatible over time, but at the same time allow the internal
representation SOPS uses to change over time. | true |
| **[spec](#sopssecretspec)** | object | SopsSecretSpec defines the desired state of SopsSecret. | false |
| **[status](#sopssecretstatus)** | object | SopsSecretStatus defines the observed state of SopsSecret. | false |


### SopsSecret.sops



Metadata is stored in SOPS encrypted files, and it contains the information necessary to decrypt the file.
This struct is just used for serialization, and SOPS uses another struct internally, sops.Metadata. It exists
in order to allow the binary format to stay backwards compatible over time, but at the same time allow the internal
representation SOPS uses to change over time.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **lastmodified** | string |  | true |
| **mac** | string |  | true |
| **[age](#sopssecretsopsageindex)** | []object |  | false |
| **[azure_kv](#sopssecretsopsazure_kvindex)** | []object |  | false |
| **encrypted_comment_regex** | string |  | false |
| **encrypted_regex** | string |  | false |
| **encrypted_suffix** | string |  | false |
| **[gcp_kms](#sopssecretsopsgcp_kmsindex)** | []object |  | false |
| **[hc_vault](#sopssecretsopshc_vaultindex)** | []object |  | false |
| **[key_groups](#sopssecretsopskey_groupsindex)** | []object |  | false |
| **[kms](#sopssecretsopskmsindex)** | []object |  | false |
| **mac_only_encrypted** | boolean |  | false |
| **[pgp](#sopssecretsopspgpindex)** | []object |  | false |
| **shamir_threshold** | integer |  | false |
| **unencrypted_comment_regex** | string |  | false |
| **unencrypted_regex** | string |  | false |
| **unencrypted_suffix** | string |  | false |
| **version** | string |  | false |


### SopsSecret.sops.age[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **enc** | string |  | true |
| **recipient** | string |  | true |


### SopsSecret.sops.azure_kv[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **name** | string |  | true |
| **vault_url** | string |  | true |
| **version** | string |  | true |


### SopsSecret.sops.gcp_kms[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **resource_id** | string |  | true |


### SopsSecret.sops.hc_vault[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **engine_path** | string |  | true |
| **key_name** | string |  | true |
| **vault_address** | string |  | true |


### SopsSecret.sops.key_groups[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[age](#sopssecretsopskey_groupsindexageindex)** | []object |  | false |
| **[azure_kv](#sopssecretsopskey_groupsindexazure_kvindex)** | []object |  | false |
| **[gcp_kms](#sopssecretsopskey_groupsindexgcp_kmsindex)** | []object |  | false |
| **[hc_vault](#sopssecretsopskey_groupsindexhc_vaultindex)** | []object |  | false |
| **[kms](#sopssecretsopskey_groupsindexkmsindex)** | []object |  | false |
| **[pgp](#sopssecretsopskey_groupsindexpgpindex)** | []object |  | false |


### SopsSecret.sops.key_groups[index].age[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **enc** | string |  | true |
| **recipient** | string |  | true |


### SopsSecret.sops.key_groups[index].azure_kv[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **name** | string |  | true |
| **vault_url** | string |  | true |
| **version** | string |  | true |


### SopsSecret.sops.key_groups[index].gcp_kms[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **resource_id** | string |  | true |


### SopsSecret.sops.key_groups[index].hc_vault[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **engine_path** | string |  | true |
| **key_name** | string |  | true |
| **vault_address** | string |  | true |


### SopsSecret.sops.key_groups[index].kms[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **arn** | string |  | true |
| **aws_profile** | string |  | true |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **context** | map[string]string |  | false |
| **role** | string |  | false |


### SopsSecret.sops.key_groups[index].pgp[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | false |
| **enc** | string |  | false |
| **fp** | string |  | false |


### SopsSecret.sops.kms[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **arn** | string |  | true |
| **aws_profile** | string |  | true |
| **created_at** | string |  | true |
| **enc** | string |  | true |
| **context** | map[string]string |  | false |
| **role** | string |  | false |


### SopsSecret.sops.pgp[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **created_at** | string |  | false |
| **enc** | string |  | false |
| **fp** | string |  | false |


### SopsSecret.spec



SopsSecretSpec defines the desired state of SopsSecret.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[secrets](#sopssecretspecsecretsindex)** | []object | Define Secrets to replicate, when secret is decrypted | true |


### SopsSecret.spec.secrets[index]



SopsSecretTemplate defines the map of secrets to create

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **name** | string | Name must be unique within a namespace. Is required when creating resources, although
some resources may allow a client to request the generation of an appropriate name
automatically. Name is primarily intended for creation idempotence and configuration
definition.
Cannot be updated.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names | true |
| **annotations** | map[string]string | Map of string keys and values that can be used to organize and categorize
(scope and select) objects. May match selectors of replication controllers
and services.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels | false |
| **data** | map[string]string | Data map to use in Kubernetes secret (equivalent to Kubernetes Secret object data, please see for more
information: https://kubernetes.io/docs/concepts/configuration/secret/#overview-of-secrets) | false |
| **immutable** | boolean | Immutable, if set to true, ensures that data stored in the Secret cannot
be updated (only object metadata can be modified).
If not set to true, the field can be modified at any time.
Defaulted to nil. | false |
| **labels** | map[string]string | Map of string keys and values that can be used to organize and categorize
(scope and select) objects. May match selectors of replication controllers
and services.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels | false |
| **stringData** | map[string]string | stringData map to use in Kubernetes secret (equivalent to Kubernetes Secret object stringData, please see for more
information: https://kubernetes.io/docs/concepts/configuration/secret/#overview-of-secrets) | false |
| **type** | enum | Kubernetes secret type.
Defaults to Opaque.
Allowed values:
- Opaque
- kubernetes.io/service-account-token
- kubernetes.io/dockercfg
- kubernetes.io/dockerconfigjson
- kubernetes.io/basic-auth
- kubernetes.io/ssh-auth
- kubernetes.io/tls
- bootstrap.kubernetes.io/token<br/><i>Enum</i>: Opaque, kubernetes.io/service-account-token, kubernetes.io/dockercfg, kubernetes.io/dockerconfigjson, kubernetes.io/basic-auth, kubernetes.io/ssh-auth, kubernetes.io/tls, bootstrap.kubernetes.io/token<br/> | false |


### SopsSecret.status



SopsSecretStatus defines the observed state of SopsSecret.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[condition](#sopssecretstatuscondition)** | object | Conditions represent the latest available observations of an instances state | false |
| **[providers](#sopssecretstatusprovidersindex)** | []object | Providers used on this secret | false |
| **[secrets](#sopssecretstatussecretsindex)** | []object | Secrets being replicated by this SopsSecret | false |
| **size** | integer | Amount of Secrets<br/><i>Default</i>: 0<br/> | false |


### SopsSecret.status.condition



Conditions represent the latest available observations of an instances state

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **lastTransitionTime** | string | lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/><i>Format</i>: date-time<br/> | true |
| **message** | string | message is a human readable message indicating details about the transition.
This may be an empty string. | true |
| **reason** | string | reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty. | true |
| **status** | enum | status of the condition, one of True, False, Unknown.<br/><i>Enum</i>: True, False, Unknown<br/> | true |
| **type** | string | type of condition in CamelCase or in foo.example.com/CamelCase. | true |
| **observedGeneration** | integer | observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/><i>Format</i>: int64<br/><i>Minimum</i>: 0<br/> | false |


### SopsSecret.status.providers[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **name** | string | Name of Object | true |
| **namespace** | string | namespace of Object | false |
| **uid** | string | namespace of Object | false |


### SopsSecret.status.secrets[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[condition](#sopssecretstatussecretsindexcondition)** | object | Condition contains details for one aspect of the current state of this API Resource. | true |
| **name** | string |  | true |
| **namespace** | string |  | true |
| **uid** | string | UID is a type that holds unique ID values, including UUIDs.  Because we
don't ONLY use UUIDs, this is an alias to string.  Being a type captures
intent and helps make sure that UIDs and names do not get conflated. | false |


### SopsSecret.status.secrets[index].condition



Condition contains details for one aspect of the current state of this API Resource.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **lastTransitionTime** | string | lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/><i>Format</i>: date-time<br/> | true |
| **message** | string | message is a human readable message indicating details about the transition.
This may be an empty string. | true |
| **reason** | string | reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty. | true |
| **status** | enum | status of the condition, one of True, False, Unknown.<br/><i>Enum</i>: True, False, Unknown<br/> | true |
| **type** | string | type of condition in CamelCase or in foo.example.com/CamelCase. | true |
| **observedGeneration** | integer | observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/><i>Format</i>: int64<br/><i>Minimum</i>: 0<br/> | false |
