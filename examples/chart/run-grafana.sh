docker run -d \
-p 3000:3000 \
-e 'GF_INSTALL_PLUGINS=marcusolsson-csv-datasource' \
--name=grafana \
-v grafana-storage:/var/lib/grafana grafana/grafana
