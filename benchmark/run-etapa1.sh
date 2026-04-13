#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
TF_DIR="$PROJECT_DIR/infra/terraform"

echo "========================================="
echo "  ETAPA 1: Go + PostgreSQL + Redis local"
echo "  (sem cache de leitura, sem LB)"
echo "========================================="

echo ""
echo "[1/6] Terraform apply..."
cd "$TF_DIR"
terraform apply -auto-approve

ECR_URL=$(terraform output -raw ecr_repository_url)
RDS_ENDPOINT=$(terraform output -raw rds_endpoint)
K6_IP=$(terraform output -raw k6_public_ip)

DB_PASSWORD=$(grep '^db_password' terraform.tfvars | cut -d'"' -f2)
HASH_SALT=$(grep '^hash_salt' terraform.tfvars | cut -d'"' -f2)
HASH_PEPPER=$(grep '^hash_pepper' terraform.tfvars | cut -d'"' -f2)

echo ""
echo "[2/6] Build e push Docker image..."
cd "$PROJECT_DIR"
REGION=us-east-1
ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin "$ACCOUNT.dkr.ecr.$REGION.amazonaws.com"
docker build --platform linux/amd64 -t encurtador .
docker tag encurtador:latest "$ECR_URL:latest"
docker push "$ECR_URL:latest"

echo ""
echo "[3/6] Aguardando EC2 K6 ficar pronta..."
echo "  IP: $K6_IP"
for i in $(seq 1 30); do
  ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 ubuntu@"$K6_IP" "echo ok" 2>/dev/null && break
  echo "  Tentativa $i..."
  sleep 10
done

echo ""
echo "[4/6] Configurando EC2 (Docker + Redis + App)..."
ssh -o StrictHostKeyChecking=no ubuntu@"$K6_IP" << REMOTE
  set -e
  echo "Aguardando Docker..."
  until sudo docker info >/dev/null 2>&1; do sleep 3; done

  echo "Subindo Redis local..."
  sudo docker rm -f local-redis 2>/dev/null || true
  sudo docker run -d --name local-redis -p 6379:6379 redis:7-alpine

  echo "Login no ECR..."
  aws ecr get-login-password --region us-east-1 2>/dev/null | sudo docker login --username AWS --password-stdin ${ECR_URL%/*} 2>/dev/null || true

  echo "Puxando imagem..."
  sudo docker pull ${ECR_URL}:latest

  echo "Subindo app..."
  sudo docker rm -f encurtador 2>/dev/null || true
  sudo docker run -d --name encurtador -p 8080:8080 \
    -e SERVER_PORT=:8080 \
    -e BASE_URL=http://localhost:8080 \
    -e ALLOWED_ORIGINS=* \
    -e "DATABASE_URL=postgres://postgres:${DB_PASSWORD}@${RDS_ENDPOINT}/encurtador?sslmode=require" \
    -e REDIS_URL=redis://172.17.0.1:6379 \
    -e REDIS_START_OFFSET=14000000 \
    -e HASH_SALT=${HASH_SALT} \
    -e HASH_PEPPER=${HASH_PEPPER} \
    -e HASH_MIN_LENGTH=6 \
    ${ECR_URL}:latest

  echo "Aguardando app..."
  for i in \$(seq 1 30); do
    curl -s http://localhost:8080/stats >/dev/null 2>&1 && echo "App rodando!" && break
    sleep 3
  done
REMOTE

echo ""
echo "[5/6] Copiando script K6..."
scp -o StrictHostKeyChecking=no "$SCRIPT_DIR/k6-test.js" ubuntu@"$K6_IP":~/k6-test.js

echo ""
echo "[6/6] Rodando K6..."
ssh ubuntu@"$K6_IP" "k6 run --env TARGET_URL=http://localhost:8080 ~/k6-test.js 2>&1" | tee "$SCRIPT_DIR/resultado-etapa1.txt"

echo ""
echo "========================================="
echo "  ETAPA 1 CONCLUIDA!"
echo "  Resultados: benchmark/resultado-etapa1.txt"
echo "  SSH: ssh ubuntu@${K6_IP}"
echo "========================================="
echo ""
echo ">>> LEMBRE: quando terminar, rode:"
echo ">>> cd infra/terraform && terraform destroy -auto-approve"
