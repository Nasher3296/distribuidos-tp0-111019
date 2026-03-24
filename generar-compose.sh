#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"
go run script-compose/compose-generator.go script-compose/types.go $1 $2
