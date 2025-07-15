resource "bytebase_setting" "semantic_types" {
  name = "settings/SEMANTIC_TYPES"

  semantic_types {
    id    = "full-mask"
    title = "Full mask"
    algorithm {
      full_mask {
        substitution = "***"
      }
    }
  }

  semantic_types {
    id    = "date-year-mask"
    title = "Date year mask"
    algorithm {
      range_mask {
        slices {
          start        = 0
          end          = 4
          substitution = "****"
        }
      }
    }
  }

  semantic_types {
    id    = "name-first-letter-only"
    title = "Name first letter only"
    algorithm {
      inner_outer_mask {
        prefix_len   = 1
        suffix_len   = 0
        substitution = "*"
        type         = "INNER"
      }
    }
  }
}