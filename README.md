# Profile Analysis Results

## Memory Profile Comparison

```
File: shortener
Type: inuse_space
Time: 2025-12-14 18:22:20 MSK
Showing nodes accounting for 1039.77kB, 40.57% of 2563.10kB total
Dropped 2 nodes (cum <= 12.82kB)
      flat  flat%   sum%        cum   cum%
  525.43kB 20.50% 20.50%   525.43kB 20.50%  github.com/go-playground/validator/v10.map.init.7
  515.19kB 20.10% 40.60%   515.19kB 20.10%  github.com/Doctor46-create/urlshort/internal/repository/database.(*urlRepository).initDeletionWorkers
    -513kB 20.01% 20.59%     -513kB 20.01%  runtime.allocm
  512.14kB 19.98% 40.57%   512.14kB 19.98%  github.com/go-chi/chi/v5.endpoints.Value (inline)
  512.06kB 19.98% 60.54%   512.06kB 19.98%  net.newFD (inline)
 -512.05kB 19.98% 40.57%  -512.05kB 19.98%  time.NewTimer
 -512.05kB 19.98% 20.59%  -512.05kB 19.98%  runtime.(*scavengerState).init
  512.05kB 19.98% 40.57%   512.05kB 19.98%  runtime.acquireSudog
         0     0% 40.57%   512.06kB 19.98%  database/sql.(*DB).Ping (inline)
         0     0% 40.57%   512.06kB 19.98%  database/sql.(*DB).PingContext
         0     0% 40.57%   512.06kB 19.98%  database/sql.(*DB).PingContext.func1
         0     0% 40.57%   512.06kB 19.98%  database/sql.(*DB).conn
         0     0% 40.57%   512.06kB 19.98%  database/sql.(*DB).retry
         0     0% 40.57%   512.06kB 19.98%  github.com/Doctor46-create/urlshort/internal/config/db.NewDatabase
         0     0% 40.57%   512.14kB 19.98%  github.com/Doctor46-create/urlshort/internal/handler.(*Handler).InitRouter
         0     0% 40.57%   515.19kB 20.10%  github.com/Doctor46-create/urlshort/internal/repository/database.NewURLRepository
         0     0% 40.57%  1539.40kB 60.06%  github.com/Doctor46-create/urlshort/internal/server.Execute
         0     0% 40.57%   512.14kB 19.98%  github.com/go-chi/chi/v5.(*Mux).Mount
         0     0% 40.57%   512.14kB 19.98%  github.com/go-chi/chi/v5.(*Mux).handle
         0     0% 40.57%   512.14kB 19.98%  github.com/go-chi/chi/v5.(*node).InsertRoute
         0     0% 40.57%   512.14kB 19.98%  github.com/go-chi/chi/v5.(*node).setEndpoint
         0     0% 40.57%   525.43kB 20.50%  github.com/go-playground/validator/v10.init
         0     0% 40.57%   512.06kB 19.98%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 40.57%   512.06kB 19.98%  github.com/jackc/pgx/v5.connect
```

Generated on Sun Dec 14 18:27:34 MSK 2025
