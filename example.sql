CREATE OR REPLACE FUNCTION public.plgo_example(text, integer)
  RETURNS text AS
'$libdir/plgo', 'plgo_example'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;
