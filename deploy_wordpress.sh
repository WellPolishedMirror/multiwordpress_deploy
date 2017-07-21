#
# Deploys database (if not already up)
# Adds new user to database
# 
function gen_pw() {
	date +%s%N | sha256sum | base64 | head -c 32
}

function container_exists() {
	docker ps -q -a -f name=$1
}

function container_up() {
	docker ps -q -f status=running -f name=$1
}

# Get database up and running
if [ ! $(container_up "test_sql") ]
then
	if [ $(container_exists "test_sql") ]
	then
		docker start "test_sql"
	else
		echo "Setting up database..."
		DB_PW=$(gen_pw)
		echo $DB_PW > db_passwd.txt
		docker run --name "test_sql" -h testdb -e MYSQL_ROOT_PASSWORD=$DB_PW -d mariadb:latest
		# Wait for database to come up
		sleep 30
	fi
fi

# Create new database and database user
DB_PW=`cat db_passwd.txt`
docker exec test_sql sh -c "mysql -uroot -p$DB_PW -e \"CREATE DATABASE pac\"" || exit 1 
docker exec test_sql sh -c "mysql -uroot -p$DB_PW -e \"GRANT ALL PRIVILEGES ON pac.* to pacuser@'%' IDENTIFIED BY 'password'\"" pac || exit 1
docker exec test_sql sh -c "mysql -uroot -p$DB_PW -e \"GRANT ALL PRIVILEGES ON pac.* to pacuser@'%' IDENTIFIED BY 'password';GRANT ALL PRIVILEGES ON pac.* to pacuser@'localhost' IDENTIFIED BY 'password';\"" pac || exit 1

# Create wordpress container
docker run --name "test-wordpress" \
           -p 5000:80 \
           --link test_sql:mysql \
           -e WORDPRESS_DB_USER="pacuser" \
           -e WORDPRESS_DB_PASSWORD="password" \
           -e WORDPRESS_DB_NAME="pac" \
           -d wordpress:latest
