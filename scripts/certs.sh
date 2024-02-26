#!/usr/bin/env bash


DAYS=3650



HELP="$(basename "$0") [OPTIONS] [SERVICE]

Stage nomad or consul certs locally from AWS Secrets Manager.
Also has action to remove staged certs.

OPTIONS:
     -h/--help  show this help text
     --region the nomad region (not needed for consul)
         --dc the datacenter
ARGS:
       SERVICE  nomad | consul (required)
 "
ARG_POS=0
SERVICE=""
DATACENTER=""
REGION=""

while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      echo "${HELP}"
      exit 0;
      ;;
    --dc)
      shift
      DATACENTER="${1}"
      shift
      ;;
    --region)
      shift
      REGION="${1}"
      shift
      ;;
    -*|--*)
      echo "Unknown flag $1"
      exit 1
      ;;
    *)
      case "${ARG_POS}" in
        0)
          SERVICE="${1}"
          ARG_POS=1
          shift
          ;;
        *)
          echo "Unknown argument: ${1}"
          echo ""
          echo "${HELP}"
          exit 1;
      esac
      ;;
  esac
done

