---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "intelcloud_instance Resource - intelcloud"
subcategory: ""
description: |-
  
---

# intelcloud_instance (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String)
- `spec` (Attributes) (see [below for nested schema](#nestedatt--spec))

### Optional

- `interfaces` (Attributes List) (see [below for nested schema](#nestedatt--interfaces))

### Read-Only

- `access_info` (Object) (see [below for nested schema](#nestedatt--access_info))
- `availability_zone` (String)
- `cloudaccount` (String)
- `id` (String) The ID of this resource.
- `ssh_proxy` (Object) (see [below for nested schema](#nestedatt--ssh_proxy))
- `status` (String)

<a id="nestedatt--spec"></a>
### Nested Schema for `spec`

Required:

- `instance_type` (String)
- `machine_image` (String)
- `ssh_public_key_names` (List of String)

Optional:

- `instance_group` (String)
- `user_data` (String)


<a id="nestedatt--interfaces"></a>
### Nested Schema for `interfaces`

Read-Only:

- `address` (String)
- `dns_name` (String)
- `gateway` (String)
- `name` (String)
- `prefix_length` (Number)
- `subnet` (String)
- `vnet` (String)


<a id="nestedatt--access_info"></a>
### Nested Schema for `access_info`

Read-Only:

- `username` (String)


<a id="nestedatt--ssh_proxy"></a>
### Nested Schema for `ssh_proxy`

Read-Only:

- `address` (String)
- `port` (Number)
- `user` (String)
