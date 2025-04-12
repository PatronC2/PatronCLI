variable "TAG" {
  default = "snapshot"
}

variable "REGISTRY" {
  default = "patronc2"
}

variable "GOOS" {
  default = "linux"
}

variable "GOARCH" {
  default = "amd64"
}

variable "TAG" {
  default = "latest"
}

target "builder-base" {
  dockerfile = "Dockerfile"
  context = "."
}

target "builder-linux-local" {
  inherits = ["builder-base"]
  args = {
    GOOS = "linux"
    GOARCH = "amd64"
    BINARY_NAME = "patron"
  }
  tags = [
    "${REGISTRY}/cli:linux-${TAG}",
    "${REGISTRY}/cli:linux-latest"
  ]
  output = ["type=local,dest=./output/linux"]
}

target "builder-windows-local" {
  inherits = ["builder-base"]
  args = {
    GOOS = "windows"
    GOARCH = "amd64"
    BINARY_NAME = "patron.exe"
  }
  tags = [
    "${REGISTRY}/cli:windows-${TAG}",
    "${REGISTRY}/cli:windows-latest"
  ]
  output = ["type=local,dest=./output/windows"]
}

target "builder-linux-release" {
  inherits = ["builder-linux-local"]
  output = ["type=registry"]
}

target "builder-windows-release" {
  inherits = ["builder-windows-local"]
  output = ["type=registry"]
}

group "local" {
  targets = ["builder-linux-local", "builder-windows-local"]
}

group "release" {
  targets = ["builder-linux-release", "builder-windows-release"]
}

group "default" {
    targets = ["local"]
}
