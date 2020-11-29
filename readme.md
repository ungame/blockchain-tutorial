# BlockChain

## Dificuldade do Bloco

No bitcoin o campo `Bits` representa é Target - ou valor de referência. 
Os mineradores devem em encontrar um número menor que o Target.

### Campos necessários

`Bits` (_decimal_):
```bash
486604799
```

`Bits` (_hexadecimal_):
```bash
1D00FFFF
```

> O número em hexadecimal é representado por 2 dígitos, ou seja cada 2 dígitos de 1D00FFFF equivale a 1 byte, formando um número de 4 bytes.

### Valor descompactado do Target

- Separar o primeiro byte do valor do `Bits` em hexadecimal e converter para decimal:

`1D` (_hexadecimal_) = `29` (_decimal_)

- Converter o valor restante:

`00FFFF` (_hexadecimal) = `65535` (_decimal_)

- Calcular o valor decompactado:

`65535` * 256<sup>(`29-3`)</sup>

```
# resultado em decimal:
26959535291011309493156476344723991336010898738574164086137773096960

# resultado em hexadecimal:
ffff0000000000000000000000000000000000000000000000000000
```

**Como calcular o resultado com Golang:**

```go
	x := big.NewInt(65535)
	y := big.NewInt(256)
	e := big.NewInt(29 - 3)
	y.Exp(y, e, nil)
	x.Mul(x, y)
    fmt.Println(x.String()) // decimal
    fmt.Printf("%x", x) // hexadecimal
```

> Agora será necessário descobrir qual a quantidade de bytes a direita do número para completar o valor de 32 bytes.

- Quantidade de dígitos no resultado:

`ffff0000000000000000000000000000000000000000000000000000` = `56` (_dígitos_)

- Bytes necessários:

`32` - `( 56 / 2 )` = `4` 

> é 56 divido por 2 pois a cada dois dígitos representa apenas 1 byte.

- Total de bytes a esquerda:

```
00000000
```

- Resultado final:

```
00000000ffff0000000000000000000000000000000000000000000000000000
```

Este número indica o mínimo de zeros a esquerda que um minerador deve encontrar para gerar um novo bloco e esse número deve ser menor que o Target.