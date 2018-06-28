# oauth2svc

Simple oauth2 server. Include ldap client

# Build

dep init

dep ensure

make linux

# Openldap Setup

docker run -d \
    -p 389:389 -p 636:636 \
    -v /data/slapd/database:/var/lib/ldap \
    -v /data/slapd/config:/etc/ldap/slapd.d \
--env LDAP_ORGANISATION="Kubernetes LDAP" \ --env LDAP_DOMAIN="k8s.com" \ --env LDAP_ADMIN_PASSWORD="password" \ --env LDAP_CONFIG_PASSWORD="password" \ --name openldap-server \ osixia/openldap:1.2.0

docker run -d \
    -p 443:443 \
    --env PHPLDAPADMIN_LDAP_HOSTS=10.20.0.19 \
    --name phpldapadmin \
    osixia/phpldapadmin:0.7.1

# Login
cn=admin,dc=k8s,dc=com
password

# User data

$ cat <<EOF > groups.ldif
dn: ou=People,dc=k8s,dc=com
ou: People
objectClass: top
objectClass: organizationalUnit
description: Parent object of all UNIX accounts

dn: ou=Groups,dc=k8s,dc=com
ou: Groups
objectClass: top
objectClass: organizationalUnit
description: Parent object of all UNIX groups

dn: cn=kubernetes,ou=Groups,dc=k8s,dc=com
cn: kubernetes
gidnumber: 100
memberuid: user1
memberuid: user2
objectclass: posixGroup
objectclass: top
EOF

$ ldapmodify -x -a -H ldap:// -D "cn=admin,dc=k8s,dc=com" -w password -f groups.ldif
adding new entry "ou=People,dc=k8s,dc=com"

adding new entry "ou=Groups,dc=k8s,dc=com"

$ cat <<EOF > users.ldif
dn: uid=user1,ou=People,dc=k8s,dc=com
cn: user1
gidnumber: 100
givenname: user1
homedirectory: /home/users/user1
loginshell: /bin/sh
objectclass: inetOrgPerson
objectclass: posixAccount
objectclass: top
objectClass: shadowAccount
objectClass: organizationalPerson
sn: user1
uid: user1
uidnumber: 1000
userpassword: user1

dn: uid=user2,ou=People,dc=k8s,dc=com
homedirectory: /home/users/user2
loginshell: /bin/sh
objectclass: inetOrgPerson
objectclass: posixAccount
objectclass: top
objectClass: shadowAccount
objectClass: organizationalPerson
cn: user2
givenname: user2
sn: user2
uid: user2
uidnumber: 1001
gidnumber: 100
userpassword: user2
EOF

$ ldapmodify -x -a -H ldap:// -D "cn=admin,dc=k8s,dc=com" -w password -f users.ldif
adding new entry "uid=user1,ou=People,dc=k8s,dc=com"

adding new entry "uid=user2,ou=People,dc=k8s,dc=com"

# Testing

http://localhost:9096/token?grant_type=client_credentials&client_id=user2&client_secret=user2&scope=read
