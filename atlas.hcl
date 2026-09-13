data "external_schema" "gorm" {
  program = ["go", "run", "./cmd/atlas-loader"]
}

env "local" {
  src = data.external_schema.gorm.url
  dev = "docker://postgres/16/dev?search_path=public"
  migration {
    dir = "file://internal/database/migrations"
  }
}
