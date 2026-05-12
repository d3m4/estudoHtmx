using EstudoHtmx.Data;

var builder = WebApplication.CreateBuilder(args);

builder.Services.Configure<SqliteOptions>(builder.Configuration);
builder.Services.AddSingleton<SqliteConnectionFactory>();
builder.Services.AddSingleton<SchemaInitializer>();

builder.Services.AddRazorPages();
builder.Services.AddAntiforgery(opts =>
{
    opts.HeaderName = "X-CSRF-TOKEN";
});

var app = builder.Build();

// aplica schema (idempotente) antes de aceitar requests
using (var scope = app.Services.CreateScope())
{
    scope.ServiceProvider
        .GetRequiredService<SchemaInitializer>()
        .EnsureSchema();
}

app.UseStaticFiles();
app.UseRouting();
app.UseAntiforgery();

app.MapRazorPages();

app.Run();
