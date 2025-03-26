#!/bin/bash

BASE_DIR=$(dirname $(dirname $(dirname $(realpath $0))))
PROGRAM_PATH="./cmd/elevator/main.go"
SERVICE_NAME="elevator"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
SCRIPT_PATH="${BASE_DIR}/scripts/service/service.sh"
LOGFILE="${BASE_DIR}/srcipts/service/elevator_service.log"

# Create the service script
cat <<EOL > $SCRIPT_PATH
#!/bin/bash

BASE_DIR="$BASE_DIR"
PROGRAM="$PROGRAM_PATH"
LOGFILE="$LOGFILE"

while true; do
    echo "Starting the elevator service..."
    cd \$BASE_DIR
    go run \$PROGRAM >> \$LOGFILE 2>&1
    echo "Elevator service crashed. Restarting in 5 seconds..."
    sleep 5
done
EOL

# Make the service script executable
chmod +x $SCRIPT_PATH

# Create the systemd service file
sudo bash -c "cat <<EOL > $SERVICE_FILE
[Unit]
Description=Elevator Service
After=network.target

[Service]
ExecStart=/usr/bin/bash $SCRIPT_PATH
Restart=always
RestartSec=5
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=elevator_service
User=$USER
Group=$(id -gn)

[Install]
WantedBy=multi-user.target
EOL"

# Reload systemd to recognize the new service
sudo systemctl daemon-reload

# Enable the service to start on boot
sudo systemctl enable $SERVICE_NAME.service

# If the service is already running, stop it
if systemctl is-active --quiet $SERVICE_NAME.service; then
    sudo systemctl stop $SERVICE_NAME.service
fi

# Start the service
sudo systemctl start $SERVICE_NAME.service

# Check the status of the service
sudo systemctl status $SERVICE_NAME.service