docker run --name reverseproxy --mount type=bind,source=/Users/maxedwards/source/lncapital/torq/docker/nginx.conf,target=/etc/nginx/nginx.conf,readonly -p 132:132 --rm nginx
