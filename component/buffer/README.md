# buffer

缓冲区复用

### 说明
先定义一个全局变量

```go
var (
	bufferPool = buffer.NewPool()
)
```

每次使用的时候从pool中获取buffer
```go
headBuf := bufferPool.GetWithSize(gf2HeadLength)
defer headBuf.Free()

headData := headBuf.Bytes()
if _, err := io.ReadFull(reader, headData); err != nil {
    return nil, err
}
```

### 注意事项

- buffer使用完毕，需要`Free()`
- buffer里获取到的`[]byte`，只能在`Free()`前使用，
- 同一个buffer的`[]byte`，可以通过`Reset()`重复使用


