# ntnu_ttk4145_elevators

## Design
To tackle the Button Contract and to ensure fault tolerance to power outages or network connection loss, we design our system to be a peer-to-peer network. Each peer knows about the other elevator's states and the requests in the system. By keeping this information consistent enough among all peers, they can calculate which orders they will handle on their own. In case of a crash, they can regain the requested information from the other peers in the system so that no call is lost.

### The Requests and Order System
We define requests as follows: Each call button of the elevator (cab and hall) is associated with one request data object. The object contains information about its Origin (hall or cab + floor) and its current Status: Unknown, Absent, Unconfirmed, or Confirmed. Initially, the status of the request is Unknown. When no user wants to be picked up or dropped off at the Origin of a request its status is Absent. As soon as a user presses a cab or hall button the request with the corresponding Origin changes its status to Unconfirmed. The unconfirmed requests are distributed to all peers/elevators. When a single peer has received unconfirmed requests from all alive peers of the same Origin, the request state is changed to Confirmed. When a request has been handled by an elevator it changes the status back to Absent. This process is akin to a Cyclic Counter approach. The diagram below illustrates this FSM.
![RequestFSM](https://github.com/user-attachments/assets/60809c5d-57c3-4112-a610-89dde222c7f7)

One peer of the system shares the state of its local elevator and the requests it knows of with all other peers at a regular, fixed interval. If no update is received after a certain time, the peer is considered dead and will not be considered in the confirmation process of one request.

The request mechanism and regular sharing of the elevator states ensure that all peers have consistent enough information to convert requests into orders. We define an order as an instruction to the elevator to execute a request. Only confirmed requests are converted into orders. As every peer shares the same information when assigning requests, they all decide on a common order distribution.

## System Structure
The project is divided into several modules which are separated by channels. Each module runs in its own routine. This enables a clean separation of responsibilities. The diagram below shows which modules exist and how they interact with each other.

![Modules](https://github.com/user-attachments/assets/86796711-9c2b-4447-bbf8-c36a1185ea02)

Key modules:
- **elevatorio**: Interface to the elevator hardware/simulator - handles button signals and elevator control
- **driver**: Manages the elevator's physical behavior and movement
- **requests**: Processes button presses and manages the request state machine
- **orders**: Assigns confirmed requests to specific elevators based on optimality
- **comms**: Handles peer-to-peer communication between elevators
- **healthmonitor**: Keeps track of which elevators are functioning in the system

The Hall Request Assigner algorithm (in the orders module) optimally distributes hall calls to elevators based on their current states and positions, minimizing wait time and ensuring efficient service.

## Repo Structure
The repo is structured according to [Go docu](https://go.dev/doc/modules/layout) and uses the project-layout from the [golang-standards team](https://github.com/golang-standards/project-layout).

## Prerequisites
- Go version 1.16 or higher
- Linux operating system (recommended, as the elevator simulator binary is built for Linux)
- If using Windows, you'll need WSL or similar to run the Linux executables
- xterm for running the simulation script (can be installed with `sudo apt install xterm` on Debian/Ubuntu)

## Running the Project
To run the project, follow these steps:
1. Clone the repository with submodules:
    ```sh
    git clone --recursive git@github.com:Lorcker/ntnu_ttk4145_elevators.git
    ```
2. Navigate to the project directory:
    ```sh
    cd ntnu_ttk4145_elevators
    ```
3. Configure the project by editing the `configs/config.json` file:
    ```json
    {
      "elevator_addr": "localhost:15657",
      "local_peer_id": 0,
      "local_port": 15444
    }
    ```
    - `elevator_addr`: Address of the elevator simulator or hardware.
    - `local_peer_id`: ID of the local elevator.
    - `local_port`: Port the local [comms] module listens to and sends broadcasts on.

4. Run the project:
    ```sh
    go run cmd/elevator/main.go -config=configs/config.json
    ```

## Using the Scripts
### `local_sim_testing.bash`
This script starts multiple instances of the simulator and the Go program in separate terminals for local testing. It takes the path to the simulator executable as an argument. The script will start two instances of the simulator and the Go program, each with different ports.

Usage:
```sh
./scripts/local_sim_testing.bash path_to_simulator_executable
```

### `configure.sh`
This script configures the elevator service to run as a systemd service. It creates a service script and a systemd service file, then enables and starts the service. This script is inted to install the software on production system to ensure redundancy.

Usage:
```sh
./scripts/service/configure.sh
```

## Notes
- The repository includes submodules, so it must be cloned with the `--recursive` flag.
- The `external` directory includes a binary that only runs on Linux machines.
- When deploying multiple elevators, ensure each has a unique ID!!!