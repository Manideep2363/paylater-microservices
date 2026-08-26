#!/bin/bash
set -euo pipefail

echo "=== USER ENV ==="
kubectl exec -n paylater deploy/user-service -- printenv | grep -E '^(DB_|SERVER_PORT|INTERNAL)' | sort

echo "=== INTERNAL CREATE ==="
AUTH_POD=$(kubectl get pod -n paylater -l app=auth-service -o jsonpath='{.items[0].metadata.name}')
EMAIL="probe$(date +%s)@example.com"
echo "EMAIL=$EMAIL"
if kubectl exec -n paylater "$AUTH_POD" -- \
  wget -qO- \
  --header='Content-Type: application/json' \
  --header='X-Internal-Token: paylater-internal-secret' \
  --post-data="{\"name\":\"Probe\",\"email\":\"${EMAIL}\",\"password\":\"password123\"}" \
  http://user-service:8082/internal/users; then
  echo
  echo "INTERNAL CREATE OK"
else
  echo
  echo "INTERNAL CREATE FAILED"
fi

echo "=== GATEWAY REGISTER ==="
curl -i -s -X POST http://localhost:8088/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Manideep","email":"mani@example.com","password":"password123"}'
echo

echo "=== USER LOGS ==="
kubectl logs deployment/user-service -n paylater --tail=15

echo "=== AUTH LOGS ==="
kubectl logs deployment/auth-service -n paylater --tail=15

echo "=== DB USERS ==="
kubectl exec -n paylater deploy/mysql -- \
  mysql -uroot -p'mani7737$$' -e "SELECT user_id,name,email,credit_limit,current_due FROM paylater_users.users;"
