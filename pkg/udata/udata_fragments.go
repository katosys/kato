package udata

//-----------------------------------------------------------------------------
// Imports:
//-----------------------------------------------------------------------------

import "os"

//-----------------------------------------------------------------------------
// Typedefs:
//-----------------------------------------------------------------------------

type filter struct {
	anyOf  []string
	noneOf []string
	allOf  []string
}

type fragment struct {
	filter
	data string
}

type fragmentSlice []fragment

//-----------------------------------------------------------------------------
// func: anyOf
//-----------------------------------------------------------------------------

func (f *fragment) anyOf(tags []string) bool {
	for _, tag := range tags {
		for _, filter := range f.filter.anyOf {
			if tag == filter {
				return true
			}
		}
	}
	return false
}

//-----------------------------------------------------------------------------
// func: noneOf
//-----------------------------------------------------------------------------

func (f *fragment) noneOf(tags []string) bool {
	for _, tag := range tags {
		for _, filter := range f.filter.noneOf {
			if tag == filter {
				return false
			}
		}
	}
	return true
}

//-----------------------------------------------------------------------------
// func: allOf
//-----------------------------------------------------------------------------

func (f *fragment) allOf(tags []string) bool {
	for _, filter := range f.filter.allOf {
		present := false
		for _, tag := range tags {
			if tag == filter {
				present = true
				break
			}
		}
		if !present {
			return false
		}
	}
	return true
}

//-----------------------------------------------------------------------------
// func: load
//-----------------------------------------------------------------------------

func (fragments *fragmentSlice) load() {

	//--------
	//-[etcd]-
	//--------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf:  []string{"master", "worker", "border"},
			noneOf: []string{"quorum"},
		},
		data: `
 etcd:
  name: "{{.HostName}}-{{.HostID}}"
 {{if .EtcdToken }} discovery: https://discovery.etcd.io/{{.EtcdToken}}{{else}} initial_cluster: "{{.EtcdServers}}"
  initial_cluster_state: "running"{{end}}
  advertise_client_urls: "http://{PRIVATE_IPV4}:2379"
  listen_client_urls: "http://127.0.0.1:2379,http://{PRIVATE_IPV4}:2379"
  proxy: on`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum"},
		},
		data: `
 etcd:
  name: "quorum-{{.HostID}}"
 {{if .EtcdToken }} discovery: https://discovery.etcd.io/{{.EtcdToken}}{{else}} initial-cluster: "{{.EtcdServers}}"
  initial_cluster_state: "new"{{end}}
  advertise_client_urls: "http://{PRIVATE_IPV4}:2379"
  initial_advertise_peer_urls: "http://{PRIVATE_IPV4}:2380"
  listen_client_urls: "http://127.0.0.1:2379,http://{PRIVATE_IPV4}:2379"
  listen_peer_urls: "http://{PRIVATE_IPV4}:2380"`,
	})

	//-------------
	//-[etc files]-
	//-------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 storage:
  files:
   - path: "/etc/hostname"
     filesystem: "root"
     mode: 0644
     contents:
      inline: {{.HostName}}-{{.HostID}}.{{.Domain}}`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - path: "/etc/hosts"
     filesystem: "root"
     mode: 0644
     contents:
      inline: |
       127.0.0.1 localhost
       {PRIVATE_IPV4} {{.HostName}}-{{.HostID}}.{{.Domain}} {{.HostName}}-{{.HostID}} marathon-lb
 {{range .Aliases}}      {PRIVATE_IPV4} {{.}}-{{$.HostID}}.{{$.Domain}} {{.}}-{{$.HostID}}
 {{end}}`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - path: "/etc/resolv.conf"
     filesystem: "root"
     mode: 0644
     contents:
      inline: |
       search {{.Domain}}
       nameserver 8.8.8.8
`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - path: "/etc/kato.env"
     filesystem: "root"
     mode: 0644
     contents:
      inline: |
       KATO_CLUSTER_ID={{.ClusterID}}
       KATO_QUORUM_COUNT={{.QuorumCount}}
       KATO_ROLES='{{range $k, $v := .Roles}}{{if $k}} {{end}}{{$v}}{{end}}'
       KATO_HOST_NAME={{.HostName}}
       KATO_HOST_ID={{.HostID}}
       KATO_ZK={{.ZkServers}}
       KATO_ETCD_ENDPOINTS={{.EtcdEndpoints}}
       KATO_SYSTEMD_UNITS='{{range $k, $v := .SystemdUnits}}{{if $k}} {{end}}{{$v}}{{end}}'
       KATO_ALERT_MANAGERS={{.AlertManagers}}
       KATO_DOMAIN=$(hostname -d)
       KATO_MESOS_DOMAIN=$(hostname -d | cut -d. -f-2).mesos
       KATO_PRI_IP={PRIVATE_IPV4}
       KATO_PUB_IP={PUBLIC_IPV4}
       KATO_QUORUM=$(({{.QuorumCount}}/2 + 1))
       KATO_VOLUMES=/var/lib/rexray/volumes
       KATO_DNS_PROVIDER={{.DNSProvider}}
       KATO_DNS_API_KEY={{.DNSApiKey}}
`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - path: "/etc/rexray/rexray.env"
     filesystem: "root"
     mode: 0644
   - path: "/etc/rexray/config.yml"
     filesystem: "root"
     mode: 0644
     contents:
      inline: |
       libstorage:
         service: {{.RexrayStorageDriver}}
       {{.RexrayStorageDriver}}:
     {{- if eq .RexrayStorageDriver "ebs" }}
         region: {{.Ec2Region}}
     {{- else if eq .RexrayStorageDriver "virtualbox" }}
         endpoint: http://{{.RexrayEndpointIP}}:18083
         volumePath: ` + os.Getenv("HOME") + `/VirtualBox Volumes
         controllerName: SATA
     {{- end}}
`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - path: "/etc/rkt/trustedkeys/prefix.d/quay.io/kato/bff313cdaa560b16a8987b8f72abf5f6799d33bc"
     filesystem: "root"
     mode: 0644
     contents:
      inline: |
       -----BEGIN PGP PUBLIC KEY BLOCK-----
       Version: GnuPG v2

       mQENBFTT6doBCACkVncI+t4HASQdnByRlXCYkwjsPqGOlgTCgenop5I6vgTqFWhQ
       PMNhtSaFdFECMt2WKQT4QGVbfVOmIH9CLV+Muqvk4iJIAn3Nh3qp/kfMhwjGaS6m
       fWN2ARFCq4RIs9tboCNQOouaD5C26/FsQtIsoqyYcdX+YFaU1a+R1kp0fc2CABDI
       k6Iq8oEJO+FOYvqQYIJNfd3c0NHICilMu2jO3yIsw80qzWoFAAblyb0zVq/hudWB
       4vdVzPmJe1f4Ymk8l1R413bN65LcbCiOax3hmFWovJoxlkL7WoGTTMfaeb2QmaPL
       qcu4Q94v1KG87gyxbkIo5uZdvMLdswQI7yQ7ABEBAAG0RFF1YXkuaW8gQUNJIENv
       bnZlcnRlciAoQUNJIGNvbnZlcnNpb24gc2lnbmluZyBrZXkpIDxzdXBwb3J0QHF1
       YXkuaW8+iQE5BBMBAgAjBQJU0+naAhsDBwsJCAcDAgEGFQgCCQoLBBYCAwECHgEC
       F4AACgkQcqv19nmdM7zKzggAjGFqy7Hcx6TCFXn53/inl5iyKrTu8cuF4K547XuZ
       12Dt8b6PgJ+b3z6UnMMTd0wXKGcfOmNeQ2R71xmVnviuo7xB5ZkZIBxHI4M/5uhK
       I6GZKr84WJS2ec7ssH2ofFQ5u1l+es9jUwW0KbAoNmES0IcdDy28xfmJpkfOn3oI
       P2Bzz4rGlIqJXEjq28Wk+qQu64kJRKYuPNXqiHncPDm+i5jMXUUN1D+pkDukp26x
       oLbpol42/jIcM3fe2AFZnflittBCHYLIHjJ51NlpSHJZmf2pQZbdyeKElN2SCNe7
       nDcol24zYIC+SX0K23w/LrLzlff4mzbO99ePt1bB9zAiVA==
       =SBoV
       -----END PGP PUBLIC KEY BLOCK-----
`,
	})

	//--------------
	//-[home files]-
	//--------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - path: "/home/core/.bashrc"
     filesystem: "root"
     mode: 0644
     user:
      name: "core"
     group:
      name: "core"
     contents:
      inline: |
       [[ $- = *i* ]] && {
         eval "$(katoctl --completion-script-bash)"
         alias ls='ls -hF --color=auto --group-directories-first'
         alias grep='grep --color=auto'
       } || shopt -s expand_aliases
       alias l='ls -l'
       alias ll='ls -la'
       alias dim='docker images'
       alias dps='docker ps'
       alias drm='docker rm -v $(docker ps -qaf status=exited)'
       alias drmi='docker rmi $(docker images -qf dangling=true)'
       alias drmv='docker volume rm $(docker volume ls -qf dangling=true)'
`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - path: "/home/core/.kato/{{.ClusterID}}.json"
     filesystem: "root"
     mode: 0644
     contents:
      inline: |
{{.KatoState | indent 7}}
`,
	})

	//-------------
	//-[opt files]-
	//-------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - path: "/opt/bin/etchosts"
     filesystem: "root"
     mode: 0755
     contents:
      inline: |
       #!/bin/bash
       source /etc/kato.env
       [ -f /etc/.hosts ] || cp /etc/hosts /etc/.hosts
       PUSH=$(awk "/${KATO_PRI_IP}/ {print \$1\" \"\$2\" \"\$3}" /etc/.hosts)
       for role in ${KATO_ROLES}; do
       etcdctl set /hosts/${role}/$(hostname -f) "${PUSH}"; done
       KEYS=$(etcdctl ls --recursive /hosts | grep ${KATO_DOMAIN} | \
       grep -v $(hostname -f) | rev | sort | rev | uniq -s 14 | sort)
       for key in ${KEYS}; do PULL+=$(etcdctl get ${key})$'\n'; done
       cat /etc/.hosts > /etc/hosts
       echo "${PULL}" >> /etc/hosts
`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - path: "/opt/bin/loopssh"
     filesystem: "root"
     mode: 0755
     contents:
      inline: |
       #!/bin/bash
       G=$(tput setaf 2); N=$(tput sgr0)
       A=$(grep $1 /etc/hosts | awk '{print $2}' | sort -u | grep -v int)
       for i in $A; do echo "${G}--[ $i ]--${N}"; ssh -o UserKnownHostsFile=/dev/null \
       -o StrictHostKeyChecking=no -o ConnectTimeout=3 $i -C "${@:2}" 2> /dev/null; done
`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf:  []string{"quorum", "master", "worker", "border"},
			noneOf: []string{"vagrant-virtualbox"},
		},
		data: `
   - path: "/opt/bin/dnspush"
     filesystem: "root"
     mode: 0755
     contents:
      inline: |
       #!/bin/bash
       source /etc/kato.env
       declare -A IP=(['ext']="${KATO_PUB_IP}" ['int']="${KATO_PRI_IP}")
       for ROLE in ${KATO_ROLES}; do for i in ext int; do
         katoctl ${KATO_DNS_PROVIDER} --api-key ${KATO_DNS_API_KEY:-none} record \
         add --zone ${i}.${KATO_DOMAIN} ${ROLE}-${KATO_HOST_ID}:A:${IP[${i}]}
       done done
`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - path: "/opt/bin/katostat"
     filesystem: "root"
     mode: 0755
     contents:
      inline: |
       #!/bin/bash
       source /etc/kato.env
       systemctl -p Id,LoadState,ActiveState,SubState show ${KATO_SYSTEMD_UNITS} | \
       awk 'BEGIN {RS="\n\n"; FS="\n";} {print $2"\t"$3"\t"$4"\t"$1}'
`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"master", "worker"},
		},
		data: `
   - path: "/opt/bin/zk-alive"
     filesystem: "root"
     mode: 0755
     contents:
      inline: |
       #!/bin/bash
       for t in {1..3}; do
         cnt=0; for i in $(seq ${1}); do
           echo ruok | ncat quorum-${i} 2181 | grep -q imok && cnt=$((cnt+1))
         done &> /dev/null; [ $cnt -ge $((${1}/2 + 1)) ] && exit 0 || sleep $((5*${t}))
       done; exit 1
`,
	})

	//---------------
	//-[filesystems]-
	//---------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"master", "worker"},
			allOf: []string{"ec2"},
		},
		data: `
  filesystems:
   - mount:
      device: /dev/xvdb
      format: ext4
      wipe_filesystem: true`,
	})

	//-----------------
	//-[systemd units]-
	//-----------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
 systemd:
  units:
   - name: "coreos-metadata.service"
     enable: true
     dropins:
      - name: "10-coreos-metadata.conf"
        contents: |
         [Service]
         ExecStartPost=/bin/bash -c '\
         source /run/metadata/coreos && sed -i \
         -e 's/{PRIVATE_IPV4}/'"$${COREOS_EC2_IPV4_LOCAL}"'/g' \
         -e 's/{PUBLIC_IPV4}/'"$${COREOS_EC2_IPV4_PUBLIC}"'/g' \
         /etc/hosts /etc/kato.env'
`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - name: "etchosts.service"
     enable: true
     contents: |
      [Unit]
      Description=Stores IPs and hostnames in etcd
      Requires=etcd-member.service
      After=etcd-member.service

      [Service]
      Type=oneshot
      ExecStart=/opt/bin/etchosts

      [Install]
      WantedBy=multi-user.target`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - name: "etchosts.timer"
     enable: true
     contents: |
      [Unit]
      Description=Run etchosts.service every 5 minutes
      Requires=etcd-member.service
      After=etcd-member.service

      [Timer]
      OnBootSec=1min
      OnUnitActiveSec=5min

      [Install]
      WantedBy=multi-user.target`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - name: "katoctl.service"
     enable: true
     contents: |
      [Unit]
      Description=Download katoctl

      [Service]
      Type=oneshot
      Environment=URL=https://github.com/katosys/kato/releases/download/v0.1.1
      ExecStart=/bin/bash -c " \
       [ -f /opt/bin/katoctl ] || { curl -sL -o /opt/bin/katoctl ${URL}/katoctl-linux-x86_64; }; \
       [ -x /opt/bin/katoctl ] || { chmod +x /opt/bin/katoctl; }"

      [Install]
      WantedBy=multi-user.target`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf:  []string{"quorum", "master", "worker", "border"},
			noneOf: []string{"vagrant-virtualbox"},
		},
		data: `
   - name: "dnspush.service"
     enable: true
     contents: |
      [Unit]
      Description=Publish DNS records
      After=katoctl.service

      [Service]
      Type=oneshot
      ExecStart=/bin/bash -c "PATH=${PATH}:/opt/bin exec /opt/bin/dnspush"

      [Install]
      WantedBy=multi-user.target`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"master", "worker"},
			allOf: []string{"ec2"},
		},
		data: `
   - name: "var-lib-mesos.mount"
     enable: true
     contents: |
      [Mount]
      What=/dev/xvdb
      Where=/var/lib/mesos
      Type=ext4

      [Install]
      RequiredBy=local-fs.target`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
   - name: "rexray.service"
     enable: true
     contents: |
      [Unit]
      Description=REX-Ray volume plugin
      Before=docker.service

      [Service]
      Restart=always
      RestartSec=10
      TimeoutStartSec=0
      KillMode=process
      EnvironmentFile=/etc/rexray/rexray.env
      Environment=REXRAY_URL=https://emccode.bintray.com/rexray/stable/0.11.1/rexray-Linux-x86_64-0.11.1.tar.gz
      Environment=DVDCLI_URL=https://emccode.bintray.com/dvdcli/stable/0.2.1/dvdcli-Linux-x86_64-0.2.1.tar.gz
      ExecStartPre=-/bin/bash -c " \
        [ -f /opt/bin/rexray ] || { curl -sL ${REXRAY_URL} | tar -xz -C /opt/bin; chown root:root /opt/bin/rexray; }; \
        [ -x /opt/bin/rexray ] || { chmod +x /opt/bin/rexray; }; [ -d /run/docker/plugins ] || { mkdir -p /run/docker/plugins; }"
      ExecStartPre=-/bin/bash -c " \
        [ -f /opt/bin/dvdcli ] || { curl -sL ${DVDCLI_URL} | tar -xz -C /opt/bin; chown root:root /opt/bin/dvdcli; }; \
        [ -x /opt/bin/dvdcli ] || { chmod +x /opt/bin/dvdcli; }"
      ExecStart=/opt/bin/rexray start -f
      ExecReload=/bin/kill -HUP $MAINPID

      [Install]
      WantedBy=multi-user.target`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"border"},
		},
		data: `
   - name: "mongodb.service"
     enable: true
     contents: |
      [Unit]
      Description=MongoDB
      After=rexray.service
      Requires=rexray.service

      [Service]
      Restart=always
      RestartSec=10
      TimeoutStartSec=0
      KillMode=mixed
      EnvironmentFile=/etc/kato.env
      Environment=IMG=mongo:3.7
      ExecStartPre=/usr/bin/rkt fetch --insecure-options=image docker://${IMG}
      ExecStartPre=/opt/bin/dvdcli mount --volumedriver rexray --volumename ${KATO_CLUSTER_ID}-pritunl-mongo
      ExecStart=/usr/bin/rkt run \
       --net=host \
       --dns=host \
       --hosts-entry=host \
       --volume volume-data-db,kind=host,source=${KATO_VOLUMES}/${KATO_CLUSTER_ID}-pritunl-mongo/data \
       docker://${IMG} -- \
       --bind_ip 127.0.0.1

      [Install]
      WantedBy=multi-user.target`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"border"},
		},
		data: `
   - name: "pritunl.service"
     enable: true
     contents: |
      [Unit]
      Description=Pritunl
      After=mongodb.service
      Requires=mongodb.service

      [Service]
      Restart=always
      RestartSec=10
      TimeoutStartSec=0
      KillMode=mixed
      LimitNOFILE=25000
      Environment=IMG=quay.io/kato/pritunl:v1.29.1609.88-1
      ExecStartPre=/usr/bin/rkt fetch ${IMG}
      ExecStart=/usr/bin/rkt run --stage1-from-dir=stage1-fly.aci \
       --net=host \
       --dns=host \
       --hosts-entry=host \
       --set-env MONGODB_URI=mongodb://127.0.0.1:27017/pritunl \
       ${IMG}

      [Install]
      WantedBy=multi-user.target`,
	})

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum"},
		},
		data: `
   - name: "zookeeper.service"
{{- if eq .ClusterState "existing" }}
     enable: false
{{- else}}
     enable: true
{{- end}}
     contents: |
      [Unit]
      Description=Zookeeper

      [Service]
      Restart=always
      RestartSec=10
      TimeoutStartSec=0
      KillMode=mixed
      EnvironmentFile=/etc/kato.env
      Environment=IMG=quay.io/kato/zookeeper:v3.4.8-4
      ExecStartPre=/usr/bin/sh -c "[ -d /var/lib/zookeeper ] || mkdir /var/lib/zookeeper"
      ExecStartPre=/usr/bin/rkt fetch ${IMG}
      ExecStart=/usr/bin/bash -c "exec rkt run \
       --net=host \
       --dns=host \
       --hosts-entry=host \
       --set-env=ZK_SERVER_ID=${KATO_HOST_ID} \
       --set-env=ZK_SERVERS=$${KATO_ZK//:2181/} \
       --set-env=ZK_CLIENT_PORT_ADDRESS=${KATO_PRI_IP} \
       --set-env=ZK_TICK_TIME=2000 \
       --set-env=ZK_INIT_LIMIT=5 \
       --set-env=ZK_SYNC_LIMIT=2 \
       --set-env=ZK_DATA_DIR=/var/lib/zookeeper \
       --set-env=ZK_CLIENT_PORT=2181 \
       --set-env=JMXDISABLE=false \
       --volume data,kind=host,source=/var/lib/zookeeper \
       --mount volume=data,target=/var/lib/zookeeper \
       ${IMG}"

      [Install]
      WantedBy=multi-user.target`,
	})

	/*//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"worker"},
			},
			data: `
	    - path: "/etc/marathon-lb/templates/HAPROXY_HTTP_FRONTEND_APPID_HEAD"
	      filesystem: "root"
	      mode: 0644`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"worker"},
			},
			data: `
	    - path: "/etc/cni/net.d/10-devel.conf"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          {
	            "name": "devel",
	            "type": "calico",
	            "ipam": {
	              "type": "calico-ipam"
	            },
	            "etcd_endpoints": "{{.EtcdEndpoints}}"
	          }
	    - path: "/etc/cni/net.d/10-prod.conf"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          {
	            "name": "prod",
	            "type": "calico",
	            "ipam": {
	              "type": "calico-ipam"
	            },
	            "etcd_endpoints": "{{.EtcdEndpoints}}"
	          }
	`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"master", "worker", "border"},
			},
			data: `
	    - path: "/etc/calico/resources.yaml"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          - apiVersion: v1
	            kind: ipPool
	            metadata:
	              cidr: {{.CalicoIPPool}}
	            spec:
	              ipip:
	                enabled: false
	              nat-outgoing: true
	              disabled: false
	          - apiVersion: v1
	            kind: hostEndpoint
	            metadata:
	              name: {{.HostName}}-{{.HostID}}
	              node: {{.HostName}}-{{.HostID}}.{{.Domain}}
	              labels:
	                endpoint: {{.HostName}}
	            spec:
	              expectedIPs:
	              - $private_ipv4
	          - apiVersion: v1
	            kind: policy
	            metadata:
	              name: {{.HostName}}
	            spec:
	              selector: endpoint == '{{.HostName}}'
	              ingress:
	      {{- if .HostTCPPorts}}
	              - action: allow
	                protocol: tcp
	                destination:
	                  ports: [{{range $k, $v := .HostTCPPorts}}{{if $k}},{{end}}"{{$v}}"{{end}}]{{end}}
	      {{- if .HostUDPPorts}}
	              - action: allow
	                protocol: udp
	                destination:
	                  ports: [{{range $k, $v := .HostUDPPorts}}{{if $k}},{{end}}"{{$v}}"{{end}}]{{end}}
	          - apiVersion: v1
	            kind: policy
	            metadata:
	              name: allow-egress
	            spec:
	              order: 0
	              egress:
	              - action: allow
	`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"quorum", "master", "worker", "border"},
			},
			data: `
	    - path: "/etc/rkt/trustedkeys/prefix.d/quay.io/calico/bff313cdaa560b16a8987b8f72abf5f6799d33bc"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          -----BEGIN PGP PUBLIC KEY BLOCK-----
	          Version: GnuPG v2

	          mQENBFTT6doBCACkVncI+t4HASQdnByRlXCYkwjsPqGOlgTCgenop5I6vgTqFWhQ
	          PMNhtSaFdFECMt2WKQT4QGVbfVOmIH9CLV+Muqvk4iJIAn3Nh3qp/kfMhwjGaS6m
	          fWN2ARFCq4RIs9tboCNQOouaD5C26/FsQtIsoqyYcdX+YFaU1a+R1kp0fc2CABDI
	          k6Iq8oEJO+FOYvqQYIJNfd3c0NHICilMu2jO3yIsw80qzWoFAAblyb0zVq/hudWB
	          4vdVzPmJe1f4Ymk8l1R413bN65LcbCiOax3hmFWovJoxlkL7WoGTTMfaeb2QmaPL
	          qcu4Q94v1KG87gyxbkIo5uZdvMLdswQI7yQ7ABEBAAG0RFF1YXkuaW8gQUNJIENv
	          bnZlcnRlciAoQUNJIGNvbnZlcnNpb24gc2lnbmluZyBrZXkpIDxzdXBwb3J0QHF1
	          YXkuaW8+iQE5BBMBAgAjBQJU0+naAhsDBwsJCAcDAgEGFQgCCQoLBBYCAwECHgEC
	          F4AACgkQcqv19nmdM7zKzggAjGFqy7Hcx6TCFXn53/inl5iyKrTu8cuF4K547XuZ
	          12Dt8b6PgJ+b3z6UnMMTd0wXKGcfOmNeQ2R71xmVnviuo7xB5ZkZIBxHI4M/5uhK
	          I6GZKr84WJS2ec7ssH2ofFQ5u1l+es9jUwW0KbAoNmES0IcdDy28xfmJpkfOn3oI
	          P2Bzz4rGlIqJXEjq28Wk+qQu64kJRKYuPNXqiHncPDm+i5jMXUUN1D+pkDukp26x
	          oLbpol42/jIcM3fe2AFZnflittBCHYLIHjJ51NlpSHJZmf2pQZbdyeKElN2SCNe7
	          nDcol24zYIC+SX0K23w/LrLzlff4mzbO99ePt1bB9zAiVA==
	          =SBoV
	          -----END PGP PUBLIC KEY BLOCK-----
	`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"quorum", "master", "worker", "border"},
				allOf: []string{"cacert"},
			},
			data: `
	    - path: "/etc/ssl/certs/{{.ClusterID}}.pem"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	{{.CaCert | indent 10}}
	`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"quorum", "master", "worker", "border"},
			},
			data: `
	    - path: "/home/core/.aws/config"
	      filesystem: "root"
	      mode: 0640
	      user:
	        name: "core"
	      group:
	        name: "core"
	      contents:
	        inline: |
	          [default]
	          region = {{.Ec2Region}}
	`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"quorum", "master", "worker", "border"},
			},
			data: `
	    - path: "/etc/ssh/sshd_config"
	      filesystem: "root"
	      mode: 0600
	      contents:
	        inline: |
	          UsePrivilegeSeparation sandbox
	          Subsystem sftp internal-sftp
	          ClientAliveInterval 180
	          UseDNS no
	          PermitRootLogin no
	          AllowUsers core
	          PasswordAuthentication no
	          ChallengeResponseAuthentication no
	`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"quorum", "master", "worker", "border"},
			},
			data: `
	    - path: "/opt/bin/awscli"
	      filesystem: "root"
	      mode: 0755
	      contents:
	        inline: |
	          #!/bin/bash
	          docker run -i --rm \
	          --net host \
	          --volume /home/core/.aws:/root/.aws:ro \
	          --volume ${PWD}:/aws \
	          quay.io/kato/awscli:v1.10.47-1 "${@}"
	`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf:  []string{"worker"},
				allOf:  []string{"cacert"},
				noneOf: []string{"vagrant-virtualbox"},
			},
			data: `
	 - path: "/opt/bin/getcerts"
	   filesystem: "root"
	   mode: 0755
	   contents:
	     inline: |
	       #!/bin/bash
	       [ -d /etc/certs ] || mkdir /etc/certs && cd /etc/certs
	       [ -f certs.tar.bz2 ] || /opt/bin/awscli s3 cp s3://{{.Domain}}/certs.tar.bz2 .
	`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"quorum", "master", "worker", "border"},
				allOf: []string{"cacert"},
			},
			data: `
	    - path: "/opt/bin/custom-ca"
	      filesystem: "root"
	      mode: 0755
	      contents:
	        inline: |
	          #!/bin/bash
	          source /etc/kato.env
	          [ -f /etc/ssl/certs/${KATO_CLUSTER_ID}.pem ] && {
	            ID=$(sed -n 2p /etc/ssl/certs/${KATO_CLUSTER_ID}.pem)
	            NU=$(grep -lir $ID /etc/ssl/certs/* | wc -l)
	            [ "$NU" -lt "2" ] && update-ca-certificates &> /dev/null
	          }; exit 0
	`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"master"},
				allOf: []string{"prometheus"},
			},
			data: `
	    - path: "/etc/alertmanager/config.yml"
	      filesystem: "root"
	      mode: 0600
	      contents:
	        inline: |
	          global:
	      {{- if .SMTPURL}}
	            smtp_smarthost: {{.SMTP.Host}}:{{.SMTP.Port}}
	            smtp_from: alertmanager@{{.Domain}}
	            smtp_auth_username: {{.SMTP.User}}
	            smtp_auth_password: {{.SMTP.Pass}}{{end}}
	      {{- if .SlackWebhook}}
	            slack_api_url: {{.SlackWebhook}}{{end}}

	          templates:
	          - '/etc/alertmanager/template/*.tmpl'

	          route:
	            group_by: ['alertname', 'cluster', 'service']
	            group_wait: 30s
	            group_interval: 5m
	            repeat_interval: 3h
	            receiver: operators

	          receivers:
	          - name: 'operators'
	      {{- if .SMTPURL}}
	            email_configs:
	      {{- if .AdminEmail}}
	            - to: '{{.AdminEmail}}'{{end}}{{end}}
	      {{- if .SlackWebhook}}
	            slack_configs:
	            - send_resolved: true
	              channel: kato{{end}}
	`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"master"},
				allOf: []string{"prometheus"},
			},
			data: `
	    - path: "/etc/prometheus/targets/prometheus.yml"
	      filesystem: "root"
	      mode: 0644
	    - path: "/etc/prometheus/prometheus.yml"
	      filesystem: "root"
	      mode: 0600
	      contents:
	        inline: |
	          global:
	          external_labels:
	            master: {{.HostID}}
	          scrape_interval: 15s
	          scrape_timeout: 10s
	          evaluation_interval: 10s

	          rule_files:
	          - /etc/prometheus/recording.rules
	          - /etc/prometheus/alerting.rules

	          alerting:
	          alert_relabel_configs:
	          - source_labels: [master]
	            action: replace
	            replacement: 'all'
	            target_label: master

	          scrape_configs:

	          - job_name: 'prometheus'
	            file_sd_configs:
	              - files:
	                - /etc/prometheus/targets/prometheus.yml

	          - job_name: 'cadvisor'
	            file_sd_configs:
	              - files:
	                - /etc/prometheus/targets/cadvisor.yml

	          - job_name: 'etcd'
	            file_sd_configs:
	              - files:
	                - /etc/prometheus/targets/etcd.yml

	          - job_name: 'node'
	            file_sd_configs:
	              - files:
	                - /etc/prometheus/targets/node.yml

	          - job_name: 'mesos'
	            file_sd_configs:
	              - files:
	                - /etc/prometheus/targets/mesos.yml

	          - job_name: 'haproxy'
	            file_sd_configs:
	              - files:
	                - /etc/prometheus/targets/haproxy.yml

	          - job_name: 'zookeeper'
	            file_sd_configs:
	              - files:
	                - /etc/prometheus/targets/zookeeper.yml
	`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"master"},
				allOf: []string{"prometheus"},
			},
			data: `
	    - path: "/etc/prometheus/alerting.rules"
	      filesystem: "root"
	      mode: 0600
	      contents:
	        inline: |
	          ALERT ScrapeDown
	            IF up == 0
	            FOR 5m
	            LABELS { severity = "page" }
	            ANNOTATIONS {
	              summary = "Scrape instance {{"{{"}} $labels.instance {{"}}"}} down",
	              description = "Job {{"{{"}} $labels.job {{"}}"}} has been down for more than 5 minutes.",
	            }
	`,
		})

		//----------------------------------

		*fragments = append(*fragments, fragment{
			filter: filter{
				anyOf: []string{"master"},
				allOf: []string{"prometheus"},
			},
			data: `
	    - path: "/etc/confd/conf.d/prom-prometheus.toml"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          [template]
	          src = "prom-prometheus.tmpl"
	          dest = "/etc/prometheus/targets/prometheus.yml"
	          keys = [ "/hosts/master" ]
	    - path: "/etc/confd/templates/prom-prometheus.tmpl"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "master" 1{{"}}"}}:9191{{"{{"}}end{{"}}"}}
	            labels:
	              role: master
	              shard: {{.HostID}}
	    - path: "/etc/confd/conf.d/prom-cadvisor.toml"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          [template]
	          src = "prom-cadvisor.tmpl"
	          dest = "/etc/prometheus/targets/cadvisor.yml"
	          keys = [
	            "/hosts/quorum",
	            "/hosts/master",
	            "/hosts/worker",
	          ]
	    - path: "/etc/confd/templates/prom-cadvisor.tmpl"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          - targets:{{"{{"}}range gets "/hosts/quorum/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "quorum" 1{{"}}"}}:4194{{"{{"}}end{{"}}"}}
	            labels:
	              role: quorum
	              shard: {{.HostID}}
	          - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "master" 1{{"}}"}}:4194{{"{{"}}end{{"}}"}}
	            labels:
	              role: master
	              shard: {{.HostID}}
	          - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "worker" 1{{"}}"}}:4194{{"{{"}}end{{"}}"}}
	            labels:
	              role: worker
	              shard: {{.HostID}}
	    - path: "/etc/confd/conf.d/prom-etcd.toml"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          [template]
	          src = "prom-etcd.tmpl"
	          dest = "/etc/prometheus/targets/etcd.yml"
	          keys = [
	            "/hosts/quorum",
	            "/hosts/master",
	            "/hosts/worker",
	          ]
	    - path: "/etc/confd/templates/prom-etcd.tmpl"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          - targets:{{"{{"}}range gets "/hosts/quorum/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "quorum" 1{{"}}"}}:2379{{"{{"}}end{{"}}"}}
	            labels:
	              role: quorum
	              shard: {{.HostID}}
	          - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "master" 1{{"}}"}}:2379{{"{{"}}end{{"}}"}}
	            labels:
	              role: master
	              shard: {{.HostID}}
	          - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "worker" 1{{"}}"}}:2379{{"{{"}}end{{"}}"}}
	            labels:
	              role: worker
	              shard: {{.HostID}}
	    - path: "/etc/confd/conf.d/prom-node.toml"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          [template]
	          src = "prom-node.tmpl"
	          dest = "/etc/prometheus/targets/node.yml"
	          keys = [
	            "/hosts/quorum",
	            "/hosts/master",
	            "/hosts/worker",
	          ]
	    - path: "/etc/confd/templates/prom-node.tmpl"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          - targets:{{"{{"}}range gets "/hosts/quorum/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "quorum" 1{{"}}"}}:9101{{"{{"}}end{{"}}"}}
	            labels:
	              role: quorum
	              shard: {{.HostID}}
	          - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "master" 1{{"}}"}}:9101{{"{{"}}end{{"}}"}}
	            labels:
	              role: master
	              shard: {{.HostID}}
	          - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "worker" 1{{"}}"}}:9101{{"{{"}}end{{"}}"}}
	            labels:
	              role: worker
	              shard: {{.HostID}}
	    - path: "/etc/confd/conf.d/prom-mesos.toml"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          [template]
	          src = "prom-mesos.tmpl"
	          dest = "/etc/prometheus/targets/mesos.yml"
	          keys = [
	            "/hosts/master",
	            "/hosts/worker",
	          ]
	    - path: "/etc/confd/templates/prom-mesos.tmpl"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          - targets:{{"{{"}}range gets "/hosts/master/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "master" 1{{"}}"}}:9104{{"{{"}}end{{"}}"}}
	            labels:
	              role: master
	              shard: {{.HostID}}
	          - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "worker" 1{{"}}"}}:9105{{"{{"}}end{{"}}"}}
	            labels:
	              role: worker
	              shard: {{.HostID}}
	    - path: "/etc/confd/conf.d/prom-haproxy.toml"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          [template]
	          src = "prom-haproxy.tmpl"
	          dest = "/etc/prometheus/targets/haproxy.yml"
	          keys = [ "/hosts/worker" ]
	    - path: "/etc/confd/templates/prom-haproxy.tmpl"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          - targets:{{"{{"}}range gets "/hosts/worker/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "worker" 1{{"}}"}}:9102{{"{{"}}end{{"}}"}}
	            labels:
	              role: worker
	              shard: {{.HostID}}
	    - path: "/etc/confd/conf.d/prom-zookeeper.toml"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          [template]
	          src = "prom-zookeeper.tmpl"
	          dest = "/etc/prometheus/targets/zookeeper.yml"
	          keys = [ "/hosts/quorum" ]
	    - path: "/etc/confd/templates/prom-zookeeper.tmpl"
	      filesystem: "root"
	      mode: 0644
	      contents:
	        inline: |
	          - targets:{{"{{"}}range gets "/hosts/quorum/*"{{"}}"}}
	            {{"{{"}}$base := base .Key{{"}}"}}- {{"{{"}}replace $base "{{.HostName}}" "quorum" 1{{"}}"}}:9103{{"{{"}}end{{"}}"}}
	            labels:
	              role: quorum
	              shard: {{.HostID}}
	`,
		})*/
}

//-----------------------------------------------------------------------------
// func: loadOld
//-----------------------------------------------------------------------------

func (fragments *fragmentSlice) loadOld() {

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"cacert"},
		},
		data: `
coreos:
 units:
  - name: "custom-ca.service"
    command: "start"
    content: |
     [Unit]
     Description=Re-hash SSL certificates
     Before=docker.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/custom-ca`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"ec2"},
		},
		data: `
  - name: "docker.service"
    drop-ins:
     - name: "20-docker-opts.conf"
       content: |
        [Service]
        Environment='DOCKER_OPTS=--registry-mirror=http://external-registry-sys.marathon:5000'`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
		},
		data: `
  - name: "kato.target"
    command: "start"
    enable: true
    content: |
     [Unit]
     Description=The Káto System
     After=kato-env.service network-online.target
     Requires=kato-env.service network-online.target

     [Install]
     WantedBy=multi-user.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"master", "worker", "border"},
		},
		data: `
  - name: "calico.service"
    enable: true
    content: |
     [Unit]
     Description=Calico per-host agent

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     EnvironmentFile=/etc/kato.env
     Environment=CNI_URL=https://github.com/projectcalico/cni-plugin/releases/download/v1.9.1
     Environment=CALICOCTL_URL=https://github.com/projectcalico/calicoctl/releases/download/v1.3.0
     Environment=CNI_PLUGINS=/var/lib/cni-plugins
     Environment=IMG=quay.io/calico/node:v1.3.0
     ExecStartPre=/usr/sbin/sysctl -w net.netfilter.nf_conntrack_max=1000000
     ExecStartPre=/usr/bin/sh -c "[ -d /var/run/calico ] || mkdir /var/run/calico"
     ExecStartPre=/usr/bin/sh -c "[ -d /var/log/calico ] || mkdir /var/log/calico"
     ExecStartPre=-/bin/bash -c " \
      [ -f ${CNI_PLUGINS}/calico ] || { curl -sL -o ${CNI_PLUGINS}/calico ${CNI_URL}/calico; }; \
      [ -x ${CNI_PLUGINS}/calico ] || { chmod +x ${CNI_PLUGINS}/calico; }"
     ExecStartPre=-/bin/bash -c " \
      [ -f ${CNI_PLUGINS}/calico-ipam ] || { curl -sL -o ${CNI_PLUGINS}/calico-ipam ${CNI_URL}/calico-ipam; }; \
      [ -x ${CNI_PLUGINS}/calico-ipam ] || { chmod +x ${CNI_PLUGINS}/calico-ipam; }"
     ExecStartPre=/bin/bash -c " \
      [ -f /opt/bin/calicoctl ] || { curl -sL -o /opt/bin/calicoctl ${CALICOCTL_URL}/calicoctl; }; \
      [ -x /opt/bin/calicoctl ] || { chmod +x /opt/bin/calicoctl; }"
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStartPre=/opt/bin/calicoctl create --skip-exists -f /etc/calico/resources.yaml
     ExecStart=/usr/bin/rkt run --stage1-from-dir=stage1-fly.aci \
      --net=host \
      --dns=host \
      --hosts-entry=host \
      --volume=run,kind=host,source=/run \
      --mount=volume=run,target=/run \
      --volume=modules,kind=host,source=/lib/modules \
      --mount=volume=modules,target=/lib/modules \
      --volume=var-run-calico,kind=host,source=/var/run/calico \
      --mount=volume=var-run-calico,target=/var/run/calico \
      --volume=var-log-calico,kind=host,source=/var/log/calico \
      --mount=volume=var-log-calico,target=/var/log/calico \
      --set-env=FELIX_LOGFILEPATH=/var/log/calico/felix.log \
      --set-env=FELIX_LOGSEVERITYFILE=WARNING \
      --set-env=FELIX_LOGSEVERITYSYS=WARNING \
      --set-env=FELIX_LOGSEVERITYSCREEN=WARNING \
      --set-env=NODENAME=${KATO_HOST_NAME}-${KATO_HOST_ID}.${KATO_DOMAIN} \
      --set-env=IP=${KATO_PRI_IP} \
      --set-env=CALICO_NETWORKING_BACKEND=bird \
      --set-env=ETCD_ENDPOINTS=${KATO_ETCD_ENDPOINTS} \
      --set-env=NO_DEFAULT_POOLS=true \
      ${IMG}

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"master"},
		},
		data: `
  - name: "mesos-master.service"
    enable: true
    content: |
     [Unit]
     Description=Mesos master
     After=zookeeper.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     LimitNOFILE=infinity
     TasksMax=infinity
     KillMode=mixed
     EnvironmentFile=/etc/kato.env
     Environment=IMG=quay.io/kato/mesos:v1.3.1-1
     ExecStartPre=/opt/bin/zk-alive ${KATO_QUORUM_COUNT}
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStartPre=/usr/bin/rkt run \
      --volume rootfs,kind=host,source=/ \
      --mount volume=rootfs,target=/media \
      ${IMG} --exec cp -- -R /opt /media
     ExecStart=/usr/bin/bash -c " \
      PATH=/opt/bin:${PATH} \
      LD_LIBRARY_PATH=/opt/lib:/lib64 \
      exec /opt/bin/mesos-master \
       --hostname=master-${KATO_HOST_ID}.${KATO_DOMAIN} \
       --cluster=${KATO_CLUSTER_ID} \
       --ip=${KATO_PRI_IP} \
       --zk=zk://${KATO_ZK}/mesos \
       --work_dir=/var/lib/mesos/master \
       --log_dir=/var/log/mesos \
       --quorum=${KATO_QUORUM}"

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"master"},
		},
		data: `
  - name: "mesos-dns.service"
    enable: true
    content: |
     [Unit]
     Description=Mesos DNS
     After=mesos-master.service
     Before=go-dnsmasq.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     EnvironmentFile=/etc/kato.env
     Environment=IMG=quay.io/kato/mesos-dns:v0.6.0-2
     ExecStartPre=/opt/bin/zk-alive ${KATO_QUORUM_COUNT}
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStart=/usr/bin/rkt run \
      --net=host \
      --dns=host \
      --hosts-entry=host \
      --set-env=MDNS_ZK=zk://${KATO_ZK}/mesos \
      --set-env=MDNS_REFRESHSECONDS=45 \
      --set-env=MDNS_LISTENER=${KATO_IP} \
      --set-env=MDNS_PORT={{.MesosDNSPort}} \
      --set-env=MDNS_HTTPON=false \
      --set-env=MDNS_TTL=45 \
      --set-env=MDNS_RESOLVERS=8.8.8.8 \
      --set-env=MDNS_DOMAIN=${KATO_MESOS_DOMAIN} \
      --set-env=MDNS_IPSOURCE=netinfo \
      ${IMG}
{{- if eq .MesosDNSPort 53 }}
     ExecStartPost=/usr/bin/sh -c ' \
       echo search marathon.${KATO_MESOS_DOMAIN} ${KATO_MESOS_DOMAIN} ${KATO_DOMAIN} > /etc/resolv.conf && \
       echo "nameserver ${KATO_PRI_IP}" >> /etc/resolv.conf'
     ExecStopPost=/usr/bin/sh -c ' \
       echo search ${KATO_DOMAIN} > /etc/resolv.conf && \
       echo "nameserver 8.8.8.8" >> /etc/resolv.conf'
{{- end}}

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"master"},
		},
		data: `
  - name: "marathon.service"
    enable: true
    content: |
     [Unit]
     Description=Marathon
     After=mesos-master.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     LimitNOFILE=8192
     EnvironmentFile=/etc/kato.env
     Environment=IMG=quay.io/kato/marathon:v1.4.8-1
     ExecStartPre=/opt/bin/zk-alive ${KATO_QUORUM_COUNT}
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStart=/usr/bin/rkt run \
      --net=host \
      --dns=host \
      --hosts-entry=host \
      --hostname master-${KATO_HOST_ID}.${KATO_DOMAIN} \
      --set-env=LIBPROCESS_IP=${KATO_PRI_IP} \
      --set-env=LIBPROCESS_PORT=9292 \
      --set-env=MESOS_NATIVE_JAVA_LIBRARY=/opt/lib/libmesos.so \
      --volume lib,kind=host,source=/opt/lib \
      --mount volume=lib,target=/opt/lib \
      ${IMG} -- \
      --no-logger \
      --checkpoint \
      --http_address ${KATO_PRI_IP} \
      --master zk://${KATO_ZK}/mesos \
      --zk zk://${KATO_ZK}/marathon \
      --task_launch_timeout 240000 \
      --hostname master-${KATO_HOST_ID}.${KATO_DOMAIN} \
      --enable_features external_volumes

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "confd.service"
    enable: true
    content: |
     [Unit]
     Description=Lightweight configuration management tool
     After=etcd2.service
     Requires=etcd2.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     Environment=IMG=quay.io/kato/confd:v0.13.0-1
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStart=/usr/bin/rkt run \
      --net=host \
      --volume etc,kind=host,source=/etc \
      --mount volume=etc,target=/etc \
      ${IMG} -- \
      -node http://127.0.0.1:2379 \
      -watch

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "alertmanager.service"
    enable: true
    content: |
     [Unit]
     Description=Alertmanager service
     Before=prometheus.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     EnvironmentFile=/etc/kato.env
     Environment=IMG=quay.io/kato/alertmanager:v0.8.0-1
     ExecStartPre=/usr/bin/sh -c "[ -d /etc/alertmanager ] || mkdir -p /etc/alertmanager"
     ExecStartPre=/usr/bin/sh -c "[ -d /var/lib/alertmanager ] || mkdir -p /var/lib/alertmanager"
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStart=/usr/bin/rkt run \
      --net=host \
      --dns=host \
      --hosts-entry=host \
      --volume volume-etc-alertmanager,kind=host,source=/etc/alertmanager,readOnly=true \
      --volume volume-var-lib-alertmanager,kind=host,source=/var/lib/alertmanager \
      ${IMG} -- \
      -log.level=info \
      -web.listen-address=${KATO_PRI_IP}:9093 \
      -web.external-url=http://master-${KATO_HOST_ID}.${KATO_DOMAIN}:9093 \
      -config.file=/etc/alertmanager/config.yml \
      -storage.path=/var/lib/alertmanager

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "prometheus.service"
    enable: true
    content: |
     [Unit]
     Description=Prometheus service
     After=rexray.service confd.service
     Requires=rexray.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     EnvironmentFile=/etc/kato.env
     Environment=IMG=quay.io/kato/prometheus:v1.7.1-1
     ExecStartPre=/usr/bin/sh -c "[ -d /etc/prometheus ] || mkdir /etc/prometheus"
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStartPre=/opt/bin/dvdcli mount --volumedriver rexray --volumename ${KATO_CLUSTER_ID}-prometheus-${KATO_HOST_ID}
     ExecStart=/usr/bin/rkt run \
      --net=host \
      --dns=host \
      --hosts-entry=host \
      --volume volume-etc-prometheus,kind=host,source=/etc/prometheus,readOnly=true \
      --volume volume-var-lib-prometheus,kind=host,source=${KATO_VOLUMES}/${KATO_CLUSTER_ID}-prometheus-${KATO_HOST_ID}/data \
      ${IMG} --exec /usr/local/bin/prometheus -- \
      -config.file=/etc/prometheus/prometheus.yml \
      -storage.local.path=/var/lib/prometheus \
      -alertmanager.url ${KATO_ALERT_MANAGERS} \
      -web.external-url=http://master-${KATO_HOST_ID}.${KATO_DOMAIN}:9191 \
      -web.console.libraries=/usr/share/prometheus/console_libraries \
      -web.console.templates=/usr/share/prometheus/consoles \
      -web.listen-address=${KATO_PRI_IP}:9191

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "rkt-api.service"
    enable: true
    content: |
     [Unit]
     Description=Rocket API service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     ExecStart=/usr/bin/rkt api-service

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "cadvisor.service"
    enable: true
    content: |
     [Unit]
     Description=cAdvisor service
     After=docker.service rkt-api.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     EnvironmentFile=/etc/kato.env
     Environment=URL=https://github.com/google/cadvisor/releases/download/v0.26.1
     ExecStartPre=/bin/bash -c " \
      [ -f /opt/bin/cadvisor ] || { curl -sL -o /opt/bin/cadvisor ${URL}/cadvisor; }; \
      [ -x /opt/bin/cadvisor ] || { chmod +x /opt/bin/cadvisor; }"
     ExecStart=/opt/bin/cadvisor \
      --listen_ip ${KATO_PRI_IP} \
      --logtostderr \
      --port=4194

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"master"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "mesos-master-exporter.service"
    enable: true
    content: |
     [Unit]
     Description=Prometheus mesos master exporter
     Wants=mesos-master.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     EnvironmentFile=/etc/kato.env
     Environment=IMG=quay.io/kato/exporters:v0.2.0-2
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStartPre=/usr/bin/systemctl is-active mesos-master.service
     ExecStart=/usr/bin/rkt run \
      --net=host \
      ${IMG} --exec mesos_exporter -- \
      -master http://${KATO_PRI_IP}:5050 \
      -addr :9104

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum", "master", "worker", "border"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "node-exporter.service"
    enable: true
    content: |
     [Unit]
     Description=Prometheus node exporter
     After=network-online.service
     Requires=network-online.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     EnvironmentFile=/etc/kato.env
     Environment=IMG=quay.io/kato/exporters:v0.2.0-2
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStart=/usr/bin/rkt run \
      --net=host \
      ${IMG} --exec node_exporter -- \
      -web.listen-address :9101

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"quorum"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "zookeeper-exporter.service"
{{- if eq .ClusterState "existing" }}
    command: "stop"
    enable: false
{{- else}}
    enable: true
{{- end}}
    content: |
     [Unit]
     Description=Prometheus zookeeper exporter
     Wants=zookeeper.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     EnvironmentFile=/etc/kato.env
     Environment=IMG=quay.io/kato/exporters:v0.2.0-2
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStartPre=/usr/bin/systemctl is-active zookeeper.service
     ExecStart=/usr/bin/sh -c "exec rkt run \
      --net=host \
      ${IMG} --exec zookeeper_exporter -- \
      -web.listen-address :9103 \
      $(echo ${KATO_ZK} | tr , ' ')"

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "go-dnsmasq.service"
    enable: true
    content: |
     [Unit]
     Description=Lightweight caching DNS proxy
     After=etchosts.timer

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     EnvironmentFile=/etc/kato.env
     Environment=IMG=quay.io/kato/go-dnsmasq:v1.0.7-1
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStartPre=/usr/bin/etcdctl ls /hosts/master
     ExecStartPre=/usr/bin/sh -c " \
       { for i in $(etcdctl ls /hosts/master); do \
       etcdctl get $${i} | awk '/master/ {print $1\":{{.MesosDNSPort}}\"}'; done \
       | tr '\n' ','; echo 8.8.8.8; } > /tmp/ns"
     ExecStart=/usr/bin/sh -c "exec rkt run \
      --net=host \
      --hosts-entry=host \
      --volume dns,kind=host,source=/etc/resolv.conf \
      --mount volume=dns,target=/etc/resolv.conf \
      ${IMG} -- \
      --listen ${KATO_PRI_IP} \
      --nameservers $(cat /tmp/ns) \
      --hostsfile /etc/hosts \
      --hostsfile-poll 60 \
      --default-resolver \
      {{range .StubZones}}--stubzones {{.}} \
      {{end -}}
      --search-domains marathon.${KATO_MESOS_DOMAIN},${KATO_MESOS_DOMAIN},${KATO_DOMAIN} \
      --enable-search"

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "mesos-agent.service"
    enable: true
    content: |
     [Unit]
     Description=Mesos agent
     After=go-dnsmasq.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     LimitNOFILE=infinity
     TasksMax=infinity
     KillMode=mixed
     EnvironmentFile=/etc/kato.env
     Environment=IMG=quay.io/kato/mesos:v1.3.1-1
     ExecStartPre=/opt/bin/zk-alive ${KATO_QUORUM_COUNT}
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStartPre=/usr/bin/rkt run \
      --volume rootfs,kind=host,source=/ \
      --mount volume=rootfs,target=/media \
      ${IMG} --exec cp -- -R /opt /media
     ExecStart=/usr/bin/bash -c " \
      PATH=/opt/bin:${PATH} \
      LD_LIBRARY_PATH=/opt/lib:/lib64 \
      exec /opt/bin/mesos-agent \
      --executor_environment_variables='{\"LD_LIBRARY_PATH\": \"/opt/lib:/lib64\"}' \
      --hostname=worker-${KATO_HOST_ID}.${KATO_DOMAIN} \
      --ip=${KATO_PRI_IP} \
      --containerizers=mesos,docker \
      --image_providers=docker \
      --docker_store_dir=/var/lib/mesos/store/docker \
      --isolation=filesystem/linux,docker/runtime,docker/volume \
      --executor_registration_timeout=5mins \
      --master=zk://${KATO_ZK}/mesos \
      --work_dir=/var/lib/mesos/agent \
      --log_dir=/var/log/mesos/agent \
      --network_cni_config_dir=/etc/cni/net.d \
      --network_cni_plugins_dir=/var/lib/cni-plugins"

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "marathon-lb.service"
    enable: true
    content: |
     [Unit]
     Description=Marathon load balancer
     After=marathon.service mesos-dns.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     Environment=IMG=mesosphere/marathon-lb:v1.10.2
     ExecStartPre=/usr/bin/rkt fetch --insecure-options=image docker://${IMG}
     ExecStartPre=/usr/bin/sh -c "until host marathon; do sleep 3; done"
     ExecStart=/usr/bin/rkt run --stage1-from-dir=stage1-fly.aci \
      --net=host \
      --dns=host \
      --hosts-entry=host \
      --set-env=PORTS=9090,9091 \
      --set-env=HAPROXY_RELOAD_SIGTERM_DELAY=5 \
      --volume templates,kind=host,source=/etc/marathon-lb/templates \
      --mount volume=templates,target=/marathon-lb/templates \
      docker://${IMG} --exec /marathon-lb/run -- sse \
      --marathon http://marathon:8080 \
      --health-check \
      --group external \
      --group internal \
      --haproxy-map

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "cni-plugins.service"
    enable: true
    content: |
     [Unit]
     Description=Get the CNI plugins
     Before=mesos-agent.service

     [Service]
     Type=oneshot
     ExecStart=/usr/bin/sh -c "[ -d /var/lib/cni-plugins ] || mkdir -p /var/lib/cni-plugins"
     ExecStart=/usr/bin/rkt run \
       --volume cni,kind=host,source=/var/lib/cni-plugins \
       --mount volume=cni,target=/tmp \
       quay.io/kato/cni-plugins:v0.6.0-1

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf:  []string{"worker"},
			allOf:  []string{"cacert"},
			noneOf: []string{"vbox"},
		},
		data: `
  - name: "getcerts.service"
    enable: true
    content: |
     [Unit]
     Description=Get certificates from private S3 bucket
     Requires=docker.service
     Before=go-dnsmasq.service
     After=docker.service

     [Service]
     Type=oneshot
     ExecStart=/opt/bin/getcerts

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "docker-gc.service"
    enable: true
    content: |
     [Unit]
     Description=Docker garbage collector
     Requires=etcd2.service docker.service
     After=etcd2.service docker.service

     [Service]
     Type=oneshot
     WorkingDirectory=/tmp
     ExecStart=/bin/bash -c '\
       docker ps -aq --no-trunc | sort -u > containers.all; \
       docker ps -q --no-trunc | sort -u > containers.running; \
       docker rm $$(comm -23 containers.all containers.running) 2>/dev/null; \
       docker rmi $$(docker images -qf dangling=true) 2>/dev/null; \
       docker volume rm $(docker volume ls -f dangling=true | awk "/^local/ {print $2}") 2>/dev/null; \
       etcdctl set /docker/images/$$(hostname) "$$(docker ps --format "{{"{{"}}.Image{{"}}"}}" | sort -u)"; \
       for i in $$(etcdctl ls /docker/images); do etcdctl get $$i; done | sort -u > images.running; \
       docker images | awk "{print \$$1\\":\\"\$$2}" | sed 1d | sort -u > images.local; \
       for i in $$(comm -23 images.local images.running | grep -v kato | grep -v mesosphere); \
       do docker rmi $$i; done; true'

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"worker"},
		},
		data: `
  - name: "docker-gc.timer"
    enable: true
    content: |
     [Unit]
     Description=Run docker-gc.service every 12 hours

     [Timer]
     OnBootSec=0s
     OnUnitActiveSec=12h

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"worker"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "haproxy-exporter.service"
    enable: true
    content: |
     [Unit]
     Description=Prometheus haproxy exporter
     Wants=marathon-lb.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     Environment=IMG=quay.io/kato/exporters:v0.2.0-2
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStartPre=/usr/bin/systemctl is-active marathon-lb.service
     ExecStart=/usr/bin/rkt run \
      --net=host \
      ${IMG} --exec haproxy_exporter -- \
      -haproxy.scrape-uri 'http://localhost:9090/haproxy?stats;csv' \
      -web.listen-address :9102

     [Install]
     WantedBy=kato.target`,
	})

	//----------------------------------

	*fragments = append(*fragments, fragment{
		filter: filter{
			anyOf: []string{"worker"},
			allOf: []string{"prometheus"},
		},
		data: `
  - name: "mesos-agent-exporter.service"
    enable: true
    content: |
     [Unit]
     Description=Prometheus mesos agent exporter
     Wants=mesos-agent.service

     [Service]
     Slice=kato.slice
     Restart=always
     RestartSec=10
     TimeoutStartSec=0
     KillMode=mixed
     EnvironmentFile=/etc/kato.env
     Environment=IMG=quay.io/kato/exporters:v0.2.0-2
     ExecStartPre=/usr/bin/rkt fetch ${IMG}
     ExecStartPre=/usr/bin/systemctl is-active mesos-agent.service
     ExecStart=/usr/bin/rkt run \
      --net=host \
      ${IMG} --exec mesos_exporter -- \
      -slave http://${KATO_PRI_IP}:5051 \
      -addr :9105

     [Install]
     WantedBy=kato.target`,
	})
}
