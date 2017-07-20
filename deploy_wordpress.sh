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
		DB_PW=$(gen_pw)
		echo $DB_PW > db_passwd.txt
		docker run --name "test_sql" -e MYSQL_ROOT_PASSWORD=$DB_PW -d mariadb:latest
	fi
fi

# TODO Create new database and database user
# TODO Create wordpress container
