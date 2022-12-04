FROM centos:centos7
ENV DOWNLOAD_LOCATION=http://wabld-automation.Ivn.broadcom.net/ISO/Agent/Agent-12.0.0WASYSAGT1200000.iso 
ENV INSTALL_LOCATION=/opt/CA/WA_Agent
# Install the agent
COPY setup//tmp/ RUN yum -y install epel-release; yum -y install wget p7zip p7zip-plugins; yum clean all;

RUN wget $DOWNLOAD_LOCATION -P /tmp; ls -1 / /tmp; chmod a+x /tmp/start.sh; mkdir /tmp/ISO; cd /tmp/ISO; pwd; 7z x /tmp/Agent-12.0.0-WASYSAGT1200000.iso, chmod -R a+x/tmp/ISO; ls -1 /tmp/ISO; /tmp/ISO/WA_System_Agent_12_0_0/linux_x86_64/setup.bin -f /tmp/unix_installer.properties -i silent; echo rc=$?; mv /tmp/start.sh $INSTALL_LOCATION/; 7z x -o$INSTALL_LOCATION/jars/ext/ /tmp/OpenShift*.zip; rm -rf /tmp/*; chmod -R g+w $INSTALL_LOCATION; cd $INSTALL_LOCATION; yes "'|./PluginInstaller addons/appservices.cfg .; yes ""|| ./PluginInstaller addons/webservice.cfg.

EXPOSE 7520 
VOLUME ["$INSTALL_LOCATION/database", "SINSTALL_LOCATION/log", "$INSTALL_LOCATION/spool"] CMD INSTALL LOCATION/start sh
