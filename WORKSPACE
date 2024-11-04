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
    sha256 = "67b4d1f517ba73e0a92eb2f57d821f2ddc21f5bc2bd7a231573f11bd8758192e",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.50.0/rules_go-v0.50.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.50.0/rules_go-v0.50.0.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "b760f7fe75173886007f7c2e616a21241208f3d90e8657dc65d36a771e916b6a",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.39.1/bazel-gazelle-v0.39.1.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.39.1/bazel-gazelle-v0.39.1.tar.gz",
    ],
)

http_archive(
    name = "rules_proto",
    sha256 = "303e86e722a520f6f326a50b41cfc16b98fe6d1955ce46642a5b7a67c11c0f5d",
    strip_prefix = "rules_proto-6.0.0",
    url = "https://github.com/bazelbuild/rules_proto/releases/download/6.0.0/rules_proto-6.0.0.tar.gz",
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

load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies")

rules_proto_dependencies()

load("@bazel_features//:deps.bzl", "bazel_features_deps")

bazel_features_deps()

load("@rules_proto//proto:toolchains.bzl", "rules_proto_toolchains")

rules_proto_toolchains()

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(version = "1.23.0")

http_archive(
    name = "rules_pkg",
    sha256 = "8a298e832762eda1830597d64fe7db58178aa84cd5926d76d5b744d6558941c2",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_pkg/releases/download/0.7.0/rules_pkg-0.7.0.tar.gz",
        "https://github.com/bazelbuild/rules_pkg/releases/download/0.7.0/rules_pkg-0.7.0.tar.gz",
    ],
)

load("@rules_pkg//:deps.bzl", "rules_pkg_dependencies")

rules_pkg_dependencies()

http_archive(
    name = "rules_license",
    sha256 = "6157e1e68378532d0241ecd15d3c45f6e5cfd98fc10846045509fb2a7cc9e381",
    urls = [
        "https://github.com/bazelbuild/rules_license/releases/download/0.0.4/rules_license-0.0.4.tar.gz",
    ],
)

load("@rules_license//:deps.bzl", "rules_license_dependencies")

rules_license_dependencies()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

go_repository(
    name = "io_opencensus_go",
    importpath = "go.opencensus.io",
    sum = "h1:y73uSU6J157QMP2kn2r30vwW1A2W2WFwSCGnAVxeaD0=",
    version = "v0.24.0",
)

go_repository(
    name = "org_golang_google_grpc",
    importpath = "google.golang.org/grpc",
    sum = "h1:bs/cUb4lp1G5iImFFd3u5ixQzweKizoZJAwBNLR42lc=",
    version = "v1.65.0",
)

go_repository(
    name = "io_opencensus_go_contrib_exporter_stackdriver",
    importpath = "contrib.go.opencensus.io/exporter/stackdriver",
    sum = "h1:lIFYmQsqejvlq+GobFUbC5F0prD5gvhP6r0gWLZRDq4=",
    version = "v0.13.8",
)

go_repository(
    name = "org_golang_google_api",
    importpath = "google.golang.org/api",
    sum = "h1:9yuVqlu2JCvcLg9p8S3fcFLZij8EPSyvODIY1rkMizQ=",
    version = "v0.103.0",
)

GOOGLEAPIS_GIT_SHA = "5e6eb76b661f7002a944add149e1ed840ca82bae"  # Feb 1, 2024

http_archive(
    name = "com_google_googleapis",
    sha256 = "3285aa5fadc1ea023a82109cf417d818e593df1698faba47ca451ea2502dc43c",
    strip_prefix = "googleapis-" + GOOGLEAPIS_GIT_SHA,
    urls = ["https://github.com/googleapis/googleapis/archive/" + GOOGLEAPIS_GIT_SHA + ".tar.gz"],
)

load("@com_google_googleapis//:repository_rules.bzl", "switched_rules_by_language")

switched_rules_by_language(
    name = "com_google_googleapis_imports",
    go = True,
    grpc = True,
)

go_repository(
    name = "com_github_googleapis_gax_go_v2",
    build_directives = [
        "gazelle:resolve proto google/rpc/code.proto @com_google_googleapis//google/rpc:code_proto",
        "gazelle:resolve proto go google/rpc/code.proto @org_golang_google_genproto_googleapis_rpc//code:code",
    ],
    importpath = "github.com/googleapis/gax-go/v2",
    sum = "h1:9V9PWXEsWnPpQhu/PeQIkS4eGzMlTLGgt80cUUI8Ki4=",
    version = "v2.11.0",
)

go_repository(
    name = "org_golang_google_genproto_googleapis_api",
    importpath = "google.golang.org/genproto/googleapis/api",
    sum = "h1:2oV8dfuIkM1Ti7DwXc0BJfnwr9csz4TDXI9EmiI+Rbw=",
    version = "v0.0.0-20241021214115-324edc3d5d38",
)

go_repository(
    name = "com_google_cloud_go_iam",
    build_file_proto_mode = "disable_global",
    importpath = "cloud.google.com/go/iam",
    sum = "h1:hlQJMovyJJwYjZcTohUH4o1L8Z8kYz+E+W/zktiLCBc=",
    version = "v1.0.0",
)

go_repository(
    name = "org_golang_google_genproto_googleapis_rpc",
    importpath = "google.golang.org/genproto/googleapis/rpc",
    sum = "h1:XSJ8Vk1SWuNr8S18z1NZSziL0CPIXLCCMDOEFtHBOFc=",
    version = "v0.0.0-20230530153820-e85fd2cbaebc",
)

go_repository(
    name = "com_google_cloud_go_compute_metadata",
    importpath = "cloud.google.com/go/compute/metadata",
    sum = "h1:mg4jlk7mCAj6xXp9UJ4fjI9VUI5rubuGBW5aJ7UnBMY=",
    version = "v0.2.3",
)

go_repository(
    name = "com_github_googleapis_enterprise_certificate_proxy",
    importpath = "github.com/googleapis/enterprise-certificate-proxy",
    sum = "h1:y8Yozv7SZtlU//QXbezB6QkpuE6jMD2/gfzk4AftXjs=",
    version = "v0.2.0",
)

go_repository(
    name = "com_google_cloud_go",
    importpath = "cloud.google.com/go",
    sum = "h1:YHLKNupSD1KqjDbQ3+LVdQ81h/UJbJyZG203cEfnQgM=",
    version = "v0.111.0",
)

go_repository(
    name = "com_google_cloud_go_monitoring",
    importpath = "cloud.google.com/go/monitoring",
    sum = "h1:c9riaGSPQ4dUKWB+M1Fl0N+iLxstMbCktdEwYSPGDvA=",
    version = "v1.8.0",
)

go_repository(
    name = "com_google_cloud_go_pubsub",
    importpath = "cloud.google.com/go/pubsub",
    sum = "h1:6SPCPvWav64tj0sVX/+npCBKhUi/UjJehy9op/V3p2g=",
    version = "v1.33.0",
)

go_repository(
    name = "com_google_cloud_go_trace",
    importpath = "cloud.google.com/go/trace",
    sum = "h1:qO9eLn2esajC9sxpqp1YKX37nXC3L4BfGnPS0Cx9dYo=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_census_instrumentation_opencensus_proto",
    build_extra_args = ["-exclude=src"],  # See https://github.com/census-instrumentation/opencensus-proto/issues/200
    importpath = "github.com/census-instrumentation/opencensus-proto",
    sum = "h1:iKLQ0xPNFxR/2hzXZMrBo8f1j86j5WHzznCCQxV/b8g=",
    version = "v0.4.1",
)

go_repository(
    name = "com_github_aws_aws_sdk_go",
    importpath = "github.com/aws/aws-sdk-go",
    sum = "h1:0xphMHGMLBrPMfxR2AmVjZKcMEESEgWF8Kru94BNByk=",
    version = "v1.27.0",
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
    sha256 = "b1e80761a8a8243d03ebca8845e9cc1ba6c82ce7c5179ce2b295cd36f7e394bf",
    urls = ["https://github.com/bazelbuild/rules_docker/releases/download/v0.25.0/rules_docker-v0.25.0.tar.gz"],
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
    digest = "sha256:89f4d2d70727ff2fcccd6674a7b6dd1456e054732718ecc9904e6a9c16cf760c",
    registry = "gcr.io",
    repository = "cloud-marketplace-containers/google/ubuntu2404",
)
