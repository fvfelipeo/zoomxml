[cite_start]O manual da NFS-e disponibiliza os seguintes endpoints para integração de sistemas, utilizando o padrão **REST** com dados **JSON**[cite: 8]:

| Nome do Serviço                  | URL                   | Método HTTP | Tipo       |
| :------------------------------- | :-------------------- | :---------- | :--------- |
| Geração de NFS-e                 | `/gerar`              | `POST`      | Síncrono   |
| Cancelamento de NFS-e            | `/cancelar`           | `POST`      | Síncrono   |
| Substituição de NFS-e            | `/substituir`         | `POST`      | Síncrono   |
| Consulta de NFS-e                | `/consultar`          | `GET`       | Síncrono   |
| Retorna o número do último RPS enviado | `/ultimorpsenviado`   | `GET`       | Síncrono   |
| Consulta XML da NFS-e            | `/xmInfse`            | `GET`       | Síncrono   |

---

## Detalhes dos Endpoints

[cite_start]Cada chamada a esses endpoints deve incluir no cabeçalho os parâmetros `Authorization: SecurityKey`, `Content-Type: application/json` e `Accept: application/json`[cite: 7].

### 1. Geração de NFS-e (`/gerar`)
[cite_start]Este serviço recebe um **Recibo Provisório de Serviço (RPS)**, realiza validações e gera a NFS-e, retornando a estrutura da nota gerada ou uma mensagem de erro[cite: 9].

#### Estrutura de Dados para Geração de NFS-e:
* [cite_start]**Dados da Nota Fiscal**: Inclui município de prestação (código IBGE), natureza da operação (tributação no município, fora, isento, imune, exigibilidade suspensa), se o ISS é retido (S/N), e observações (até 255 caracteres)[cite: 10].
* [cite_start]**Atividade do Serviço Prestado**: Contém o código da atividade (fornecido pela Prefeitura, setor de ISS), **Código CNAE 2.0** (somente números), e **Código LC-116** (somente números, sem zeros à esquerda)[cite: 11, 12].
* [cite_start]**Prestador do Serviço**: Requer a inscrição municipal do prestador[cite: 11].
* [cite_start]**Tomador do Serviço**: Abrange tipo de pessoa (F-Física, J-Jurídica, E-Exterior), número do documento (CPF, CNPJ, NIF), razão social, endereço completo (logradouro, número, complemento, bairro, município, CEP), contato (telefone, e-mail), inscrição estadual e municipal[cite: 11, 12].
* [cite_start]**Dados do RPS**: Informações como número sequencial, série, tipo (Recibo Provisório de Serviços, RPS Nota Fiscal Conjugada, Cupom), e data de emissão no formato `AAAA-MM-DD`[cite: 12, 13].
* [cite_start]**Serviços Prestados**: Permite múltiplos serviços, cada um com unidade de medida, quantidade, descrição (limitada pelo tamanho da discriminação de 2000 caracteres, que inclui todas as informações do serviço e observação da NFS-e), e valor unitário[cite: 13].
* [cite_start]**Valores da Nota**: Detalha valores totais dos serviços, deduções, outras deduções, retenções (PIS, COFINS, INSS, IR, CSLL), outras retenções, descontos (incondicionado e condicionado), base de cálculo, alíquota do serviço (com 4 casas decimais), valor do ISS, valor do crédito (não obrigatório), e valor total da nota[cite: 13].

### 2. Cancelamento de NFS-e (`/cancelar`)
Este serviço permite **cancelar uma NFS-e** já emitida, sem vincular a um RPS ou a uma NFS-e substituta. [cite_start]Caso a nota não exista ou já esteja cancelada, uma mensagem de retorno será enviada[cite: 14]. [cite_start]Para o cancelamento, é necessário informar o número da NFS-e, o motivo do cancelamento e a inscrição municipal do prestador[cite: 14].

### 3. Substituição de NFS-e (`/substituir`)
[cite_start]O manual lista este endpoint, mas não fornece detalhes adicionais na seção de "Serviços Disponibilizados" sobre sua estrutura de dados, apenas menciona-o como um serviço síncrono que utiliza o método `POST`[cite: 8].

### 4. Consulta de NFS-e (`/consultar`)
Este serviço permite a consulta de NFS-e, retornando informações como o "Código de Validação" e o "LinkNfse" no JSON de retorno. O campo "DataEmissao" inclui hora, minutos e segundos. [cite_start]Um novo parâmetro de consulta é `NumeroRps`[cite: 3, 8].

### 5. Retorna o número do último RPS enviado (`/ultimorpsenviado`)
[cite_start]Este é um serviço **novo** introduzido no manual, com o objetivo de recuperar o número do último RPS enviado[cite: 3, 8].

### 6. Consulta XML da NFS-e (`/xmInfse`)
Este serviço permite **retornar o XML das notas fiscais** de serviço emitidas em lote. O XML é disponibilizado no padrão **ZIP/Base64**. [cite_start]As requisições são paginadas, retornando no máximo 100 registros por página, e o parâmetro `nr_page` pode ser usado para buscar páginas adicionais[cite: 380]. [cite_start]Os parâmetros de consulta incluem: `nr_inicial` (número inicial da NFS-e), `nr_final` (número final), `dt_inicial` (período de emissão inicial), `dt_final` (período de emissão final), `nr_competencia` (competência da NFS-e no formato AAAA MM), e `nr_page` (número da página)[cite: 380]. [cite_start]Além disso, um novo campo chamado `XmlCompactado` foi incluído no JSON de retorno da consulta da NFS-e[cite: 3].

---
### Ambientes de API
O sistema disponibiliza ambientes de **Homologação** e **Produção**.

* [cite_start]**Homologação**: O EndPoint para a API é `https://api-nfse-homologacao.prefeituramoderna.com.br/ws/services/nome_do_servico` e a SecurityKey é `9f16d93554dc1d93656e23bd4fc9d4566a4d76848517634d7bcabd5731e4948f`[cite: 7].
* [cite_start]**Produção**: O EndPoint é `https://api-nfse-nomedomunicipio-uf.prefeituramoderna.com.br/ws/services/nome_do_servico`, e a Security Key é fornecida pela Administração Tributária Municipal, podendo ser verificada no portal de produção na opção "Token API"[cite: 7].