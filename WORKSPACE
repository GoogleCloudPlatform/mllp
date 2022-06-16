# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "ab21448cef298740765f33a7f5acee0607203e4ea321219f2a4c85a6e0fb0a27",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.32.0/rules_go-v0.32.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.32.0/rules_go-v0.32.0.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "62ca106be173579c0a167deb23358fdfe71ffa1e4cfdddf5582af26520f1c66f",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.23.0/bazel-gazelle-v0.23.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.23.0/bazel-gazelle-v0.23.0.tar.gz",
    ],
)

http_archive(
    name = "com_google_protobuf",
    sha256 = "d0f5f605d0d656007ce6c8b5a82df3037e1d8fe8b121ed42e536f569dec16113",
    strip_prefix = "protobuf-3.14.0",
    urls = [
        "https://mirror.bazel.build/github.com/protocolbuffers/protobuf/archive/v3.14.0.tar.gz",
        "https://github.com/protocolbuffers/protobuf/archive/v3.14.0.tar.gz",
    ],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains(version = "1.18.2")

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

go_repository(
    name = "io_opencensus_go",
    importpath = "go.opencensus.io",
    sum = "h1:gqCw0LfLxScz8irSi8exQc7fyQ0fKQU/qnC/X8+V/1M=",
    version = "v0.23.0",
)

go_repository(
    name = "org_golang_google_grpc",
    importpath = "google.golang.org/grpc",
    sum = "h1:c+E5hkHV2oYLLcjZ0Uulu4thvOFKB0a9TWvowIWqgu4=",
    version = "v1.39.0-dev.0.20210518002758-2713b77e8526",
)

go_repository(
    name = "io_opencensus_go_contrib_exporter_stackdriver",
    sum = "h1:lIFYmQsqejvlq+GobFUbC5F0prD5gvhP6r0gWLZRDq4=",
    version = "v0.13.8",
    importpath = "contrib.go.opencensus.io/exporter/stackdriver",
)

go_repository(
   name = "org_golang_google_api",
    importpath = "google.golang.org/api",
    sum = "h1:MDkAbYIB1JpSgCTOCYYoIec/coMlKK4oVbpnBLLcyT0=",
    version = "v0.58.0",
    build_file_proto_mode = "disable_global",
)

go_repository(
    name = "com_github_googleapis_gax_go",
    commit = "be11bb253a768098254dc71e95d1a81ced778de3",
    importpath = "github.com/googleapis/gax-go",
)

go_repository(
    name = "com_google_cloud_go",
    importpath = "cloud.google.com/go",
    sum = "h1:y/cM2iqGgGi5D5DQZl6D9STN/3dR/Vx5Mp8s752oJTY=",
    version = "v0.99.0",
)

go_repository(
    name = "com_google_cloud_go_monitoring",
    importpath = "cloud.google.com/go/monitoring",
    sum = "h1:BbbME861YCj/jJnvO/gVcPmqqjfGhiGgFu3DFeP09yU=",
    version = "v1.0.0",
)

go_repository(
    name = "com_google_cloud_go_pubsub",
    importpath = "cloud.google.com/go/pubsub",
    sum = "h1:ukjixP1wl0LpnZ6LWtZJ0mX5tBmjp1f8Sqer8Z2OMUU=",
    version = "v1.3.1",
)

go_repository(
    name = "com_google_cloud_go_trace",
    importpath = "cloud.google.com/go/trace",
    sum = "h1:laKx2y7IWMjguCe5zZx6n7qLtREk4kyE69SXVC0VSN8=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_census_instrumentation_opencensus_proto",
    importpath = "github.com/census-instrumentation/opencensus-proto",
    sum = "h1:glEXhBS5PSLLv4IXzLA5yPRVX4bilULVyxxbrfOtDAk=",
    version = "v0.2.1",
    build_extra_args = ["-exclude=src"],  # See https://github.com/census-instrumentation/opencensus-proto/issues/200
)

go_repository(
    name = "com_github_aws_aws_sdk_go",
    importpath = "github.com/aws/aws-sdk-go",
    sum = "h1:m45+Ru/wA+73cOZXiEGLDH2d9uLN3iHqMc0/z4noDXE=",
    version = "v1.15.11",
)

go_repository(
    name = "com_github_go_ini_ini",
    importpath = "github.com/go-ini/ini",
    sum = "h1:Mujh4R/dH6YL8bxuISne3xX2+qcQ9p0IxKAP6ExWoUo=",
    version = "v1.25.4",
)

go_repository(
   name = "com_github_jmespath_go_jmespath",
    importpath = "github.com/jmespath/go-jmespath",
    sum = "h1:SMvOWPJCES2GdFracYbBQh93GXac8fq7HeN6JnpduB8=",
    version = "v0.0.0-20160803190731-bd40a432e4c7",
)

go_repository(
    name = "org_golang_x_oauth2",
    importpath = "golang.org/x/oauth2",
    sum = "h1:B333XXssMuKQeBwiNODx4TupZy7bf4sxFZnN2ZOcvUE=",
    version = "v0.0.0-20211005180243-6b3c2da341f1",
)

go_repository(
    name = "org_golang_x_sync",
    commit = "43a5402ce75a95522677f77c619865d66b8c57ab",
    importpath = "golang.org/x/sync",
)

go_repository(
    name = "org_golang_x_net",
    importpath = "golang.org/x/net",
    sum = "h1:qOfNqBm5gk93LjGZo1MJaKY6Bph39zOKz1Hz2ogHj1w=",
    version = "v0.0.0-20211011170408-caeb26a5c8c0",
)

go_repository(
    name = "com_github_kylelemons_godebug",
    commit = "fa7b53cdfc9105c70f134574002f406232921437",
    importpath = "github.com/kylelemons/godebug",
)

go_repository(
    name = "com_github_golang_groupcache",
    importpath = "github.com/golang/groupcache",
    sum = "h1:ZgQEtGgCBiWRM39fZuwSd1LwSqqSW0hOdXCYYDX0R3I=",
    version = "v0.0.0-20190702054246-869f871628b6",
)

go_repository(
    name = "com_github_golang_glog",
    commit = "23def4e6c14b4da8ac2ed8007337bc5eb5007998",
    importpath = "github.com/golang/glog",
)

go_repository(
    name = "org_golang_x_text",
    commit = "06d492aade888ab8698aad35476286b7b555c961",
    importpath = "golang.org/x/text",
)

go_repository(
    name = "com_github_google_go_cmp",
    commit = "cb8c7f84fcfb230736f1e5922b3132f47bc88500",
    importpath = "github.com/google/go-cmp",
)

go_repository(
    name = "com_github_google_uuid",
    commit = "0e4e31197428a347842d152773b4cace4645ca25",
    importpath = "github.com/google/uuid",
)
http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "27d53c1d646fc9537a70427ad7b034734d08a9c38924cc6357cc973fed300820",
    strip_prefix = "rules_docker-0.24.0",
    urls = ["https://github.com/bazelbuild/rules_docker/releases/download/v0.24.0/rules_docker-v0.24.0.tar.gz"],
)

load("@io_bazel_rules_docker//go:image.bzl", _go_image_repos = "repositories")

_go_image_repos()

load("@io_bazel_rules_docker//repositories:repositories.bzl", container_repositories = "repositories")

container_repositories()

load("@io_bazel_rules_docker//repositories:deps.bzl", container_deps = "deps")

container_deps()

load("@io_bazel_rules_docker//container:container.bzl", "container_pull")

container_pull(
    name = "ubuntu",
    registry = "gcr.io",
    repository = "cloud-marketplace-containers/google/ubuntu1604",
    digest = "sha256:8f0b64fd212007183434b8b3271b723700ab14e4230b5bec1415b79aaa3ac97b",
)