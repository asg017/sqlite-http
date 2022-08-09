import suncalc from "https://cdn.skypack.dev/suncalc";
import { serve } from "https://deno.land/std@0.148.0/http/server.ts";

const port = 3001;

const suntimesRoute = new URLPattern({ pathname: "/suncalc" });

interface SuntimePostParams {
  latitude: number;
  longitude: number;
  date: string;
}

// user POSTs JSON object SuntimePostParams to "/suntimes", return suncalc.getTimes response
async function handleSuntimes(request: Request): Promise<Response> {
  const { latitude, longitude, date } =
    (await request.json()) as SuntimePostParams;
  const times = suncalc.getTimes(new Date(date), latitude, longitude);
  return Promise.resolve(new Response(JSON.stringify(times), { status: 200 }));
}

await serve(
  (request: Request): Promise<Response> => {
    if (suntimesRoute.test(request.url)) return handleSuntimes(request);
    return Promise.resolve(new Response("not found :(", { status: 404 }));
  },
  { port }
);
