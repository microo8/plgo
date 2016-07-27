/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

CREATE OR REPLACE FUNCTION public.plgo_example(text, integer)
  RETURNS text AS
'$libdir/plgo', 'plgo_example'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;
