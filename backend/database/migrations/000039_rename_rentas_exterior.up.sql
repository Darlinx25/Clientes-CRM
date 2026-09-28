-- Renombra el tipo de empresa "Rentas Exterior" a "RtasFinExter" para
-- unificar el nombre de la app con el usado en las planillas de importación.
UPDATE company_types SET name = 'RtasFinExter', updated_at = datetime('now')
    WHERE name = 'Rentas Exterior';