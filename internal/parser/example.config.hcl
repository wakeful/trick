select_profile = profile.simple

# -region eu-west-1 -refresh 5 \
# -role arn::42::role-a -role arn::42::role-b -role arn::42::role-c
profile "simple" {
  region = "eu-west-1"

  chain {
    ttl = 5

    use {
      arn = "arn::42::role-a"
    }

    use {
      arn = "arn::42::role-b"
    }

    use {
      arn = "arn::42::role-c"
    }
  }
}

# -region eu-west-1 -refresh 15 \
# -role arn::42::role-a -role arn::42::role-b \
# -role arn::42::role-c -role arn::42::role-d \
# -use  arn::42::role-a -use  arn::42::role-d
profile "complex" {
  region = "eu-west-1"

  chain {
    ttl = 15

    use {
      arn  = "arn::42::role-a"
      skip = false # Defaults to false; you can skip it.
    }

    use {
      arn  = "arn::42::role-b"
      skip = true
    }

    use {
      arn  = "arn::42::role-c"
      skip = true
    }

    use {
      arn = "arn::42::role-d"
    }
  }
}

# -role arn::42::role-a -role arn::42::role-b \
# -role arn::42::role-c -role arn::42::role-d \
# -use  arn::42::role-a -use  arn::42::role-d
profile "with_defaults" {
  chain {
    use {
      arn  = "arn::42::role-a"
      skip = false # Defaults to false; you can skip it.
    }

    use {
      arn  = "arn::42::role-b"
      skip = true
    }

    use {
      arn  = "arn::42::role-c"
      skip = true
    }

    use {
      arn = "arn::42::role-d"
    }
  }
}
