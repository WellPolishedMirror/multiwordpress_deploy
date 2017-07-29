#
# Deploys database (if not already up)
# Adds new user to database
# 

# Configurations
web_name="web_net"
db_name="web_db"
if [ -z $1 ]
then
	echo "Usafe: deploy_wordpress NAME"
	exit 1
else
	wp_name=$1
fi
nginx_name="reverse_proxy"

# Helper functions
function gen_pw() {
	date +%s%N | sha256sum | base64 | head -c 32
}

function container_exists() {
	docker ps -q -a -f name=$1
}

function container_up() {
	docker ps -q -f status=running -f name=$1
}

#
# Main script 
#

# Set up network if not already there
docker network inspect $web_name &> /dev/null || docker network create web_net

# Get database up and running
if [ ! $(container_up $db_name) ]
then
	if [ $(container_exists $db_name) ]
	then
		docker start $db_name
	else
		echo "Setting up database..."
		DB_PW=$(gen_pw)
		echo $DB_PW > db_passwd.txt
		docker run --name $db_name --net $web_name -e MYSQL_ROOT_PASSWORD=$DB_PW -d mariadb:latest
	fi
	# Wait for database to come up
	sleep 30
fi

# Create new database and database user
DB_PW=`cat db_passwd.txt`
wp_user=${wp_name%%.*}
wp_pass=$(gen_pw)
echo $wp_pass > "$wp_user".txt
docker exec $db_name sh -c "mysql -uroot -p$DB_PW -e \"CREATE DATABASE ${wp_user}\"" || exit 1 
docker exec $db_name sh -c "mysql -uroot -p$DB_PW -e \"GRANT ALL PRIVILEGES ON ${wp_user}.* to $wp_user@'%' IDENTIFIED BY '$wp_pass'\"" pac || exit 1
docker exec $db_name sh -c "mysql -uroot -p$DB_PW -e \"GRANT ALL PRIVILEGES ON ${wp_user}.* to $wp_user@'%' IDENTIFIED BY '$wp_pass';GRANT ALL PRIVILEGES ON ${wp_user}.* to $wp_user@'localhost' IDENTIFIED BY '$wp_pass';\"" pac || exit 1

# Create wordpress container
echo "Setting up wordpress image ${wp_name}..."
docker run --name "$wp_name" \
           --net $web_name \
           -e WORDPRESS_DB_HOST="$db_name" \
           -e WORDPRESS_DB_USER="$wp_user" \
           -e WORDPRESS_DB_PASSWORD="$wp_pass" \
           -e WORDPRESS_DB_NAME="$wp_user" \
           -d wordpress:latest

# Set up nginx
if [ ! $(container_up "$nginx_name") ]
then
	if [ $(container_exists "$nginx_name") ]
	then
		docker start "$nginx_name"
	else
		echo "Setting up reverse proxy..."
		docker run --name $nginx_name \
                           -v $PWD/nginx-conf:/etc/nginx/:ro \
                           --net $web_name \
                           -p 80:80 \
                           -d nginx
       fi
	# Wait for nginx container to come up
	sleep 30
fi

# generate nginx configuration
printf "server {
	listen 80;
	listen [::]:80;
	server_name %s;
	location / {
	proxy_pass http://%s;
	}
}" $wp_name $wp_name > $PWD/nginx-conf/sites-enabled/$wp_name
# reload nginx configuration
docker exec -t $nginx_name nginx -s reload
