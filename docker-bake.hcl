variable "TAG" {
  default = "snapshot"
}

variable "COMMIT" {
  default = "unknown"
}

variable "BUILD_DATE" {
  default = "unknown"
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

target "builder-base" {
  dockerfile = "Dockerfile"
  context = "."
}

target "builder-linux-local" {
  inherits = ["builder-base"]
  args = {
    GOOS        = "linux"
    GOARCH      = "amd64"
    BINARY_NAME = "patron"
    TAG         = "${TAG}" 
    COMMIT      = "${COMMIT}"
    BUILD_DATE  = "${BUILD_DATE}"
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
    TAG         = "${TAG}" 
    COMMIT      = "${COMMIT}"
    BUILD_DATE  = "${BUILD_DATE}"
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
