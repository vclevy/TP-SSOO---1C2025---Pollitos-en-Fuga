## Checkpoint

Para cada checkpoint de control obligatorio, se debe crear un tag en el
repositorio con el siguiente formato:

```
checkpoint-{número}
```

Donde `{número}` es el número del checkpoint.

Para crear un tag y subirlo al repositorio, podemos utilizar los siguientes
comandos:

```bash
git tag -a checkpoint-{número} -m "Checkpoint {número}"
git push origin checkpoint-{número}
```

Asegúrense de que el código compila y cumple con los requisitos del checkpoint
antes de subir el tag.

---
---

# Instrucciones de Ejecución del Proyecto

Este documento detalla los pasos necesarios para ejecutar los distintos módulos del sistema.

---

## 1. Encender Kernel

Ejecutar el proceso principal del Kernel, indicando la configuración, el archivo de pseudocódigo y el tamaño total de memoria a usar.

```bash
go run kernel/kernel.go config/config.json <ARCHIVO.TXT> <TAMAÑO>
```

- `<ARCHIVO.TXT>`: Ruta del archivo con el pseudocódigo de procesos.
- `<TAMAÑO>`: Tamaño total de memoria en bytes o páginas (según lo que interprete tu implementación).

---

## 2. Encender Memoria

Ejecutar el proceso que gestiona la Memoria Principal y el espacio SWAP.

```bash
go run memoria/memoria.go
```

> Asegurate de que el archivo de configuración `config.json` tenga la ruta correcta del archivo SWAP y parámetros como tamaño de marco y cantidad de marcos.

---

## 3. Encender CPU

Levantar una CPU virtual con un nombre identificador.

```bash
go run cpu/cpu.go <NOMBRE>
```

- `<NOMBRE>`: Identificador único para la CPU (por ejemplo, `CPU1`, `CPU2`, etc.).

---

## 4. Encender IO

Iniciar un módulo de entrada/salida con nombre propio.

```bash
go run io/io.go <NOMBRE>
```

- `<NOMBRE>`: Nombre del dispositivo IO (por ejemplo, `DISCO`, `TECLADO`, etc.).

---

## Consideraciones

- Todos los módulos deben ejecutarse desde la raíz del proyecto.
- El archivo `config.json` debe estar correctamente configurado antes de iniciar los módulos.
- Los procesos se comunicarán entre sí por HTTP o sockets, según lo defina tu implementación.

