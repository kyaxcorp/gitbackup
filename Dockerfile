# syntax=docker/dockerfile:1

FROM golang:1.21 as go-build

WORKDIR /app
COPY . /app

RUN go mod download
RUN go build -o /tmp/main

FROM rockylinux:9

RUN dnf update -y
RUN dnf install epel-release -y
RUN dnf install -y p7zip p7zip-plugins git

WORKDIR /app
COPY --from=go-build /tmp/main ./gitbackup
RUN ln -s /app/gitbackup /usr/bin/gitbackup
COPY container.entrypoint.sh ./
RUN chmod +x container.entrypoint.sh

ARG BUILD_NO=1
LABEL BUILD_NO=$BUILD_NO
ENV BUILD_NO=$BUILD_NO

ARG BUILD_VERSION=1.0.0
LABEL BUILD_VERSION=$BUILD_VERSION
ENV BUILD_VERSION=$BUILD_VERSION

ARG BUILD_STAGE=dev
LABEL BUILD_STAGE=$BUILD_STAGE
ENV BUILD_STAGE=$BUILD_STAGE

ARG BUILD_TEAMCITY_BUILD_URL=dev
LABEL BUILD_TEAMCITY_BUILD_URL=$BUILD_TEAMCITY_BUILD_URL
ENV BUILD_TEAMCITY_BUILD_URL=$BUILD_TEAMCITY_BUILD_URL

ARG BUILD_DATE
LABEL BUILD_DATE=$BUILD_DATE
ENV BUILD_DATE=$BUILD_DATE

ARG BUILD_DATE_UNIX
LABEL BUILD_DATE_UNIX=$BUILD_DATE_UNIX
ENV BUILD_DATE_UNIX=$BUILD_DATE_UNIX

ARG BUILD_VCS_COMMIT_AUTHOR
LABEL BUILD_VCS_COMMIT_AUTHOR=$BUILD_VCS_COMMIT_AUTHOR
ENV BUILD_VCS_COMMIT_AUTHOR=$BUILD_VCS_COMMIT_AUTHOR

ARG BUILD_VCS_REPO_URL
LABEL BUILD_VCS_REPO_URL=$BUILD_VCS_REPO_URL
ENV BUILD_VCS_REPO_URL=$BUILD_VCS_REPO_URL

ARG BUILD_VCS_TAG_LABEL
LABEL BUILD_VCS_TAG_LABEL=$BUILD_VCS_TAG_LABEL
ENV BUILD_VCS_TAG_LABEL=$BUILD_VCS_TAG_LABEL

ARG BUILD_VCS_TAG_LABEL_URL
LABEL BUILD_VCS_TAG_LABEL_URL=$BUILD_VCS_TAG_LABEL_URL
ENV BUILD_VCS_TAG_LABEL_URL=$BUILD_VCS_TAG_LABEL_URL

ARG BUILD_VCS_BRANCH
LABEL BUILD_VCS_BRANCH=$BUILD_VCS_BRANCH
ENV BUILD_VCS_BRANCH=$BUILD_VCS_BRANCH

ARG BUILD_VCS_COMMIT_DATE
LABEL BUILD_VCS_COMMIT_DATE=$BUILD_VCS_COMMIT_DATE
ENV BUILD_VCS_COMMIT_DATE=$BUILD_VCS_COMMIT_DATE

ARG BUILD_VCS_COMMIT_ID
LABEL BUILD_VCS_COMMIT_ID=$BUILD_VCS_COMMIT_ID
ENV BUILD_VCS_COMMIT_ID=$BUILD_VCS_COMMIT_ID

ARG BUILD_VCS_COMMIT_ID_URL
LABEL BUILD_VCS_COMMIT_ID_URL=$BUILD_VCS_COMMIT_ID_URL
ENV BUILD_VCS_COMMIT_ID_URL=$BUILD_VCS_COMMIT_ID_URL


ENTRYPOINT ["/app/container.entrypoint.sh"]