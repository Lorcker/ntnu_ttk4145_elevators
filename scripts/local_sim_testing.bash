#!/bin/bash

# Script to start multiple instances of the simulator and the Go program
# in separate terminals for local testing. The script takes the path to the
# simulator executable as an argument. The script will start two instances
# of the simulator and the Go program, each with different ports. The script
# will create configuration files for each instance and start the simulator
# and the Go program in separate terminals. The script will store the PIDs
# of the simulator and the Go program instances and will kill them when
# the user presses Enter. The script will also clean up the configuration
# files created.

# !IMPORTANT! The script uses xterm to open new terminals. Make sure xterm is
# installed on your system. If you are using a different terminal, replace
# xterm with the command to open a new terminal in the script.

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 path_to_simulator_executable"
    exit 1
fi

SIMULATOR_EXECUTABLE=$1
ELEVATOR_PROGRAM="../cmd/elevator/main.go"

# Define base ports for the simulator and the Go program
SIMULATOR_BASE_PORT=5000
GO_PORT=6000

# Define configuration templates
CONFIG_TEMPLATE='{
    "elevator_addr": "localhost:%d",
    "num_floors": 4,
    "local_peer_id": %d,
    "local_port": %d
}'

# Store PIDs of simulator and Go program instances
SIMULATOR_PIDS=()
GO_PIDS=()
GO_STATUS=()

# Function to clean up processes
cleanup() {
    echo "Killing all instances..."
    for pid in "${SIMULATOR_PIDS[@]}"; do
        kill $pid
    done
    for pid in "${GO_PIDS[@]}"; do
        kill $pid
    done
    rm config_*.json
    exit 0
}

# Trap Ctrl+C to kill all instances and cleanup
trap cleanup INT

# Function to start a new elevator instance
start_elevator_instance() {
    local i=$1
    
    SIMULATOR_PORT=$((SIMULATOR_BASE_PORT + i))
    CONFIG_FILE="config_$i.json"

    # Create configuration file
    printf "$CONFIG_TEMPLATE" $SIMULATOR_PORT $i $GO_PORT > $CONFIG_FILE

    # Calculate positions
    Y_OFFSET=$(( (i - 1) * 300 ))

    # Start the simulator in a new terminal
    xterm -hold -geometry 50x20+0+$Y_OFFSET -e "\"$SIMULATOR_EXECUTABLE\" --port $SIMULATOR_PORT" &
    
    SIMULATOR_PID=$!
    echo "Started simulator on port $SIMULATOR_PORT with PID $SIMULATOR_PID"

    # Start the Go program in a new terminal
    xterm -hold -geometry 150x20+400+$Y_OFFSET -e "go run \"$ELEVATOR_PROGRAM\" -config=\"$CONFIG_FILE\"" &

    GO_PID=$!
    echo "Started Go program instance $i with PID $GO_PID and config $CONFIG_FILE"

    # Store PIDs and status to kill and restart them later if needed
    SIMULATOR_PIDS+=($SIMULATOR_PID)
    GO_PIDS+=($GO_PID)
    GO_STATUS+=(1) # 1 means running, 0 means stopped
}

# Function to toggle the Go program instance
toggle_go_instance() {
    local i=$1
    if [ ${GO_STATUS[$i]} -eq 1 ]; then
        kill ${GO_PIDS[$i]}
        GO_STATUS[$i]=0
        echo "Stopped Go program instance $((i+1)) with PID ${GO_PIDS[$i]}"
    else
        CONFIG_FILE="config_$((i+1)).json"
        Y_OFFSET=$(( i * 300 ))
        xterm -hold -geometry 150x20+400+$Y_OFFSET -e "go run \"$ELEVATOR_PROGRAM\" -config=\"$CONFIG_FILE\"" &
        GO_PIDS[$i]=$!
        GO_STATUS[$i]=1
        echo "Restarted Go program instance $((i+1)) with PID ${GO_PIDS[$i]}"
    fi
}

# Create initial instances
for i in {1..2}; do
    start_elevator_instance $i
done

# Wait for user input to add new elevators, kill all instances, or toggle Go program instances
while true; do
    read -n 1 -s key
    if [ "$key" = "e" ]; then
        i=$(( ${#SIMULATOR_PIDS[@]} + 1 ))
        start_elevator_instance $i
    elif [ "$key" = "" ]; then
        cleanup
    elif [[ "$key" =~ [0-9] ]]; then
        i=$((key - 1))
        if [ $i -lt ${#GO_PIDS[@]} ]; then
            toggle_go_instance $i
        fi
    fi
done