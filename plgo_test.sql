CREATE OR REPLACE FUNCTION public.plgo_test()
  RETURNS void AS
'$libdir/plgo_test', 'plgo_test'
  LANGUAGE c IMMUTABLE STRICT;
