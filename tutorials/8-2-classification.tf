resource "bytebase_setting" "classification" {
  name = "settings/DATA_CLASSIFICATION"

  classification {
    id    = "classification-example"
    title = "Classification Example"

    levels {
      title = "Level 1"
      level = 1
    }
    levels {
      title = "Level 2"
      level = 2
    }

    classifications {
      id    = "1"
      title = "Basic"
    }

    classifications {
      id    = "1-1"
      title = "User basic"
      level = 1
    }

    classifications {
      id    = "1-2"
      title = "User contact info"
      level = 2
    }

    classifications {
      id    = "2"
      title = "Employment"
    }

    classifications {
      id    = "2-1"
      title = "Employment info"
      level = 2
    }
  }
}
