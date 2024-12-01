#!/usr/bin/env bash

CERT_PREFIX=""
SAVE_TO=""
EXP_DAYS=365
COUNTRY="US"
STATE="CA"
ORG="Engineering"
ORG_UNIT="Engineering Lab"
COMMON_NAME="Engineering Lab CA"
SUB_ALT_DNS=""
SUBJECT=""
ARG_POS=0

HELP="Usage: ${0} [OPTIONS] PREFIX SAVETO

ARGUMENTS:
  PREFIX  prefix is prepended generated certs files
  SAVETO  directory to save certs
      
OPTIONS:
   --exp/-e  expiration in days
    --co/-c  country
    --st/-s  state
      --org  organization
 --org-unit  organization unit
       --cn  common name
  --alt-dns  subjectAlt dns
  
"

while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      echo "${HELP}"
      exit 0;
      ;;
    --exp|-e)
      shift
      EXP_DAYS=${1}
      shift
      ;;
    --cn|-c)
      shift
      COMMON_NAME="${1}"
      shift
      ;;
    --co|-c)
      shift
      COUNTRY="${1}"
      shift
      ;;
    --st|-s)
      shift
      STATE="${1}"
      shift
      ;;
    --org)
      shift
      REGION="${1}"
      shift
      ;;
    --org-unit)
      shift
      REGION="${1}"
      shift
      ;;
    --cn)
      shift
      REGION="${1}"
      shift
      ;;
    --alt-dns)
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
          CERT_PREFIX="${1}"
          ARG_POS=1
          shift
          ;;
        1)
          SAVE_TO="${1}"
          ARG_POS=2
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

