---
PLAN: "fix: dejar de importar io en encode y decode"
EXECUTOR: jules
REVIEWER: none
---

> Este plan se despacha con el flujo CodeJob. Ver skill: agents-workflow.

# Plan — `tinywasm/json`: cerrar la puerta a `io`

## El problema

`encode.go` y `decode.go` importan `io` de la biblioteca estándar para los
conmutadores de tipo de sus destinos y orígenes:

```go
// encode.go — output: *[]byte | *string | io.Writer
case io.Writer:

// decode.go — input: []byte | string | io.Reader
```

`io` importa `errors`, y `errors` importa `internal/reflectlite`.

## Cuánto vale — la parte honesta

**Hoy, cero bytes.** Medido: quitar el import no reduce el binario, porque el
runtime de TinyGo enlaza `internal/reflectlite` de todos modos para sus
aserciones de tipo.

Se hace por la razón estructural, que sí está medida en un Worker real:

> **El impuesto de la stdlib lo cobra la última puerta que quede abierta.**

`unicode` y `bytes` entraban a ese binario por dos caminos. Cerrar uno rendía
2.119 bytes; cerrar los dos, 93.733. Cada puerta que queda abierta hace que
cerrar las demás parezca inútil, y así es como una cadena de 90 KB sobrevive
años. `tinywasm/json` está en casi todos los binarios del ecosistema: es una de
las puertas que hay que cerrar para que las demás cuenten.

## El cambio

En `encode.go`:

```go
// Writer es io.Writer redeclarado aquí. Estructuralmente idéntico, así que
// cualquier io.Writer lo satisface sin que el llamador cambie una línea.
type Writer interface {
	Write(p []byte) (n int, err error)
}
```

En `decode.go`:

```go
// Reader es io.Reader redeclarado aquí, por la misma razón que Writer.
type Reader interface {
	Read(p []byte) (n int, err error)
}
```

Sustituye `io.Writer` por `Writer` y `io.Reader` por `Reader` en los conmutadores
de tipo, en las firmas y en los comentarios de documentación que los nombran, y
borra el import.

**Esto no rompe a nadie**: Go satisface interfaces estructuralmente, así que
quien hoy pasa un `*os.File` o un `http.ResponseWriter` sigue compilando igual.

Actualiza también el mensaje de error, que hoy nombra el tipo de la stdlib:

```go
return fmt.Err("json", "encode", "output must be *[]byte, *string, or json.Writer")
```

## Criterios de aceptación

- [ ] `GOOS=js GOARCH=wasm go list -f '{{join .Imports " "}}' .` devuelve
      exactamente `github.com/tinywasm/fmt github.com/tinywasm/model unsafe`.
- [ ] `grep -rn '"io"\|"errors"\|"bytes"\|"strings"\|"strconv"\|"encoding/json"' *.go | grep -v _test`
      → vacío.
- [ ] Un test pasa un `*bytes.Buffer` de la stdlib a `Encode` y un
      `*bytes.Reader` a `Decode` —desde un `_test.go`, donde la stdlib es
      legítima— y ambos funcionan sin cambios en el llamador.
- [ ] La batería actual del repositorio pasa sin modificarse.

**Anti-footgun:** el directorio `benchmarks/` de este repositorio compila con Go
estándar y usa `io`, `bufio` y `net/http` con toda legitimidad: son programas de
comparación, no código del paquete. **No toques sus imports.**

## Fuera de alcance

Cualquier cambio en el formato producido, en el analizador o en la API pública.
Este plan sustituye dos tipos de parámetro y nada más.
