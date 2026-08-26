#!/bin/bash
set -euo pipefail

MYSQL_POD=$(kubectl get pod -n paylater -l app=mysql -o jsonpath='{.items[0].metadata.name}')
echo "MYSQL_POD=$MYSQL_POD"

kubectl cp /mnt/f/Pay_Later/paylater-microservices/k8s/database/init.sql \
  "paylater/${MYSQL_POD}:/tmp/init.sql"

kubectl exec -n paylater "$MYSQL_POD" -- \
  mysql -uroot -p'mani7737$$' -e "source /tmp/init.sql"

echo "=== VERIFY TABLES ==="
kubectl exec -n paylater "$MYSQL_POD" -- \
  mysql -uroot -p'mani7737$$' -e "
SHOW TABLES FROM paylater_users;
SHOW TABLES FROM paylater_merchants;
SHOW TABLES FROM paylater_ledger;
"
