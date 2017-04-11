CREATE OR REPLACE FUNCTION public.plgo_test()
  RETURNS void AS
'$libdir/plgo_test', 'PLGoTest'
  LANGUAGE c IMMUTABLE STRICT;
