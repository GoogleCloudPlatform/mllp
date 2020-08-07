# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "7b9bbe3ea1fccb46dcfa6c3f3e29ba7ec740d8733370e21cdc8937467b4a4349",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.22.4/rules_go-v0.22.4.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.22.4/rules_go-v0.22.4.tar.gz",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "d8c45ee70ec39a57e7a05e5027c32b1576cc7f16d9dd37135b0eddde45cf1b10",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/bazel-gazelle/releases/download/v0.20.0/bazel-gazelle-v0.20.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.20.0/bazel-gazelle-v0.20.0.tar.gz",
    ],
)

http_archive(
    name = "com_google_protobuf",
    sha256 = "9748c0d90e54ea09e5e75fb7fac16edce15d2028d4356f32211cfa3c0e956564",
    strip_prefix = "protobuf-3.11.4",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/v3.11.4.zip"],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

go_repository(
    name = "io_opencensus_go",
    commit = "46dfec7deb6e8c5d4a46f355c0da7c6d6dc59ba4",
    importpath = "go.opencensus.io",
)

go_repository(
    name = "org_golang_google_grpc",
    commit = "f67e7c03dcbcfc81906f302a87ab0d400738877d",
    importpath = "google.golang.org/grpc",
)

go_repository(
    name = "org_golang_google_api",
    commit = "c8cf5cff125e500044b60204f230024dcc49c3a3",
    importpath = "google.golang.org/api",
)

go_repository(
    name = "org_golang_google_genproto",
    commit = "fb6d0575620bc914112d1a67d55ba8c090b1aa11",
    importpath = "google.golang.org/genproto",
)

go_repository(
    name = "com_github_googleapis_gax_go",
    commit = "be11bb253a768098254dc71e95d1a81ced778de3",
    importpath = "github.com/googleapis/gax-go",
)

go_repository(
    name = "com_google_cloud_go",
    commit = "6416cc735b86e81cff3f558da87aee244e204472",
    importpath = "cloud.google.com/go",
)

go_repository(
    name = "org_golang_x_oauth2",
    commit = "bf48bf16ab8d622ce64ec6ce98d2c98f916b6303",
    importpath = "golang.org/x/oauth2",
)

go_repository(
    name = "org_golang_x_sync",
    commit = "43a5402ce75a95522677f77c619865d66b8c57ab",
    importpath = "golang.org/x/sync",
)

go_repository(
    name = "org_golang_x_net",
    commit = "d3edc9973b7eb1fb302b0ff2c62357091cea9a30",
    importpath = "golang.org/x/net",
)

go_repository(
    name = "com_github_kylelemons_godebug",
    commit = "fa7b53cdfc9105c70f134574002f406232921437",
    importpath = "github.com/kylelemons/godebug",
)

go_repository(
    name = "com_github_golang_groupcache",
    commit = "8c9f03a8e57eb486e42badaed3fb287da51807ba",
    importpath = "github.com/golang/groupcache",
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
    sha256 = "4521794f0fba2e20f3bf15846ab5e01d5332e587e9ce81629c7f96c793bb7036",
    strip_prefix = "rules_docker-0.14.4",
    urls = ["https://github.com/bazelbuild/rules_docker/releases/download/v0.14.4/rules_docker-v0.14.4.tar.gz"],
)

load("@io_bazel_rules_docker//go:image.bzl", _go_image_repos = "repositories")

_go_image_repos()

load("@io_bazel_rules_docker//repositories:repositories.bzl", container_repositories = "repositories")

container_repositories()

load("@io_bazel_rules_docker//repositories:deps.bzl", container_deps = "deps")

container_deps()

load("@io_bazel_rules_docker//repositories:pip_repositories.bzl", "pip_deps")

pip_deps()

load("@io_bazel_rules_docker//container:container.bzl", "container_pull")

container_pull(
    name = "ubuntu",
    registry = "gcr.io",
    repository = "cloud-marketplace-containers/google/ubuntu16_04",
    digest = "sha256:c81e8f6bcbab8818fdbe2df6d367990ab55d85b4dab300931a53ba5d082f4296",
)
