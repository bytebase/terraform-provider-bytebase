resource "bytebase_group" "developers" {
  email       = "developers@example.com"
  title       = "Developer Team"
  description = "Group for all developers"
  
  members {
    member = "users/${bytebase_user.dev1.email}"
    role   = "OWNER"
  }
  
  members {
    member = "users/${bytebase_user.dev2.email}"
    role   = "MEMBER"
  }
  
  members {
    member = "users/${bytebase_user.dev3.email}"
    role   = "MEMBER"
  }
}

resource "bytebase_group" "qa" {
  email       = "qa@example.com"
  title       = "QA Team"
  description = "Group for all QA testers"
  
  members {
    member = "users/${bytebase_user.qa1.email}"
    role   = "OWNER"
  }
  
  members {
    member = "users/${bytebase_user.qa2.email}"
    role   = "MEMBER"
  }
}
