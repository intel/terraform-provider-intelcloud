---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "intelcloud_object_storage_bucket Resource - intelcloud"
subcategory: ""
description: |-
  
---

# intelcloud_object_storage_bucket (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String)
- `versioned` (Boolean)

### Optional

- `security_groups` (Attributes List) (see [below for nested schema](#nestedatt--security_groups))

### Read-Only

- `cloudaccount` (String)
- `id` (String) The ID of this resource.
- `private_endpoint` (String)
- `size` (String)
- `status` (String)

<a id="nestedatt--security_groups"></a>
### Nested Schema for `security_groups`

Read-Only:

- `gateway` (String)
- `prefix_length` (Number)
- `subnet` (String)
