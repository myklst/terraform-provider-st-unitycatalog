resource "st-unitycatalog_user" "user" {
  schemas      = ["urn:ietf:params:scim:schemas:core:2.0:User"]
  display_name = "myucuser"
  emails = [{
    value   = "myucuser@mytestuc.com"
    primary = true
  }]
  active = true
}
