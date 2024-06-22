# genTypes

GenTypes is a CLI for generating typescript types from go types

## Installation

```shell
go install github.com/Embiggenerd/genTypes
```

## Usage
```shell
genTypes -s types.go -t types.d.ts
```

## Expectation

_types.go_

```go
ype ErrorData struct {
	StatusCode int       `json:"status_code,omitempty"`
	Message    string    `json:"message,omitempty"`
	Public     bool      `json:"public"`
	CreatedAt  time.Time `genType:"Date"`
	Deleted    bool
}

type Errors []ErrorData

```

_types.d.ts_

```typescript
export interface ErrorData {
  status_code?: number /* int */;
  message?: string;
  public: boolean;
  CreatedAt: Date;
  Deleted: boolean;
}

export type Errors = ErrorData[];
```