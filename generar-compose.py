import sys


HEADER = """
name: tp0
services:
"""

SERVER = """
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    volumes:
      - ./server/config.ini:/config.ini
    environment:
      - PYTHONUNBUFFERED=1
    networks:
      - testing_net
"""

CLIENT = """
  client{client_id}:
    container_name: client{client_id}
    image: client:latest
    entrypoint: /client
    volumes:
      - ./client/config.yaml:/config.yaml
    environment:
      - CLI_ID={client_id}
    networks:
      - testing_net
    depends_on:
      - server
"""

NETWORKS = """
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

OUTPUT_FILE_PARAMETER_INDEX = 1
CLIENT_INSTANCES_PARAMETER_INDEX = 2
NUMBER_OF_PARAMETERS = 3


def main():
    if len(sys.argv) != NUMBER_OF_PARAMETERS:
        print(f"Invalid number of arguments: {sys.argv}")
        print("Usage: python3 script.py <output_file> <number_of_clients>")
        return

    if not sys.argv[CLIENT_INSTANCES_PARAMETER_INDEX].isdigit():
        print(
            f"Invalid number of clients: {sys.argv[CLIENT_INSTANCES_PARAMETER_INDEX]}"
        )
        return

    client_instances = int(sys.argv[CLIENT_INSTANCES_PARAMETER_INDEX])
    output_file = sys.argv[OUTPUT_FILE_PARAMETER_INDEX]

    print(f"Generating docker-compose file with {client_instances} clients...")
    print(f"Output file: {output_file}")

    generate_docker_compose_file(client_instances, output_file)

    print("Docker-compose file generated successfully.")


def generate_docker_compose_file(client_instances: int, output_file: str):
    with open(output_file, "w") as f:
        print("Writing header...")
        f.write(HEADER)
        print("Adding server configuration...")
        f.write(SERVER)
        for i in range(1, client_instances + 1):
            print(f"Adding client {i} configuration...")
            f.write(CLIENT.format(client_id=i))
        print("Adding networks configuration...")
        f.write(NETWORKS)


if "__main__" == __name__:
    main()
