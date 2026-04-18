# PLAN: Análisis — ¿Invertir el comportamiento por defecto de OmitEmpty?

## Pregunta

¿Debería `tinywasm/json` omitir campos vacíos **por defecto** (como hace `encoding/json` con `omitempty`), requiriendo una etiqueta explícita solo cuando un campo **nunca** debe omitirse?

## Contexto actual

El comportamiento actual de `tinywasm/json`:
- Por defecto: los campos vacíos **se incluyen** en el JSON output
- Con `OmitEmpty: true` en el Schema(): los campos vacíos se omiten

Ejemplo output actual sin OmitEmpty:
```json
{"jsonrpc":"2.0","id":"1","result":{"protocolVersion":"2024-11-05"},"error":""}
```

Con OmitEmpty en `error`:
```json
{"jsonrpc":"2.0","id":"1","result":{"protocolVersion":"2024-11-05"}}
```

## ¿Afecta al tamaño del binario WASM?

**No.** Los struct tags (`json:",omitempty"`) no van al binario WASM — TinyGo no incluye metadata de reflexión. El Schema() generado por `ormc` sí está en el binario, pero `OmitEmpty bool` es 1 byte por campo. Con 52 campos en `tinywasm/mcp`, el overhead total es < 52 bytes — irrelevante.

El impacto real de `omitempty` es en el **tamaño del JSON output** (menos bytes por red/memoria), no en el binario compilado.

## Datos: uso actual de OmitEmpty en tinywasm/mcp

| Categoría | Campos totales | Con OmitEmpty |
|-----------|---------------|---------------|
| tinywasm/mcp model_orm.go | 52 | 15 (29%) |

Solo 29% de campos usan OmitEmpty — no es mayoría. Invertir el default obligaría a marcar el 71% restante con una nueva etiqueta explícita, generando más ruido del que elimina.

## Análisis del trade-off

### Opción A: mantener el default actual (incluir vacíos)
- **Pro:** Comportamiento predecible y explícito — el desarrollador decide campo por campo
- **Pro:** Compatible con el principio de tinywasm/orm: "More Explicit Code"
- **Pro:** 29% de campos usan OmitEmpty — no hay un patrón dominante que justifique el cambio
- **Con:** Requiere declarar `OmitEmpty: true` en cada campo que lo necesite

### Opción B: omitir vacíos por defecto (invertir)
- **Pro:** Reduce boilerplate cuando la mayoría de campos son omitempty
- **Pro:** Output JSON más compacto por defecto
- **Con:** Rompe compatibilidad con todos los Schema() existentes
- **Con:** Requiere nueva etiqueta (`json:",always"` o `json:",required"`) para campos que siempre deben incluirse
- **Con:** Con solo 29% de campos siendo omitempty en mcp, generaría más etiquetas, no menos
- **Con:** Comportamiento implícito — difícil de depurar cuando un campo desaparece del output inesperadamente

## Decisión: NO invertir el default

**Justificación:** El diseño de tinywasm prioriza explicitness sobre conveniencia implícita (ver [WHY.md en tinywasm/orm](../../orm/docs/WHY.md)). Con solo 29% de campos usando OmitEmpty en el uso real, invertir el default generaría más etiquetas obligatorias y más confusión, no menos.

El comportamiento actual es correcto. No se requiere cambio en `tinywasm/json` ni en `tinywasm/orm`.

---

# PLAN: Impacto de `fmt.RawJSON` en tinywasm/json

## Contexto

Ver plan principal en [tinywasm/fmt/docs/PLAN_RAW_JSON_TYPE.md](../../fmt/docs/PLAN_RAW_JSON_TYPE.md).

## Impacto en este paquete

**Ningún cambio requerido en tinywasm/json.**

El encoder ya maneja `FieldRaw` a través del Schema() generado. El flujo es:

```
fmt.RawJSON (tipo fuente) → ormc → FieldRaw en Schema() → tinywasm/json encoder
```

tinywasm/json nunca ve el tipo `RawJSON` — solo ve `FieldRaw` en el Schema en tiempo de ejecución.
El cambio es transparente para este paquete.
