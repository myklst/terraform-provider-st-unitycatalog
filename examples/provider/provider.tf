terraform {
  required_providers {
    st-unitycatalog = {
      version = "~> 0.1"
      source  = "myklst/st-unitycatalog"
    }
  }
}

provider "st-unitycatalog" {
  host = "http://localhost:8080/unitycatalog"
}
