<?php

namespace App\Services\Sped;

use DOMDocument;
use DOMElement;
use RuntimeException;

class ResponseMapper
{
    /**
     * @return array<string, mixed>
     */
    public function mapDistDFeResponse(string $xml, bool $includeXml = true): array
    {
        $dom = $this->loadXml($xml);
        $ret = $dom->getElementsByTagName('retDistDFeInt')->item(0);

        if (! $ret instanceof DOMElement) {
            throw new RuntimeException('Resposta DistDFe inválida.');
        }

        $result = [
            'cstat' => $this->nodeValue($ret, 'cStat'),
            'xmotivo' => $this->nodeValue($ret, 'xMotivo'),
            'tp_amb' => $this->nodeValue($ret, 'tpAmb'),
            'ver_aplic' => $this->nodeValue($ret, 'verAplic'),
            'dh_resp' => $this->nodeValue($ret, 'dhResp'),
            'ult_nsu' => $this->nodeValue($ret, 'ultNSU'),
            'max_nsu' => $this->nodeValue($ret, 'maxNSU'),
            'documents' => [],
            'raw_xml' => $xml,
        ];

        $lote = $ret->getElementsByTagName('loteDistDFeInt')->item(0);

        if ($lote instanceof DOMElement) {
            foreach ($lote->getElementsByTagName('docZip') as $docZip) {
                if (! $docZip instanceof DOMElement) {
                    continue;
                }

                $schema = $docZip->getAttribute('schema');
                $encoded = trim($docZip->textContent);

                $document = [
                    'nsu' => $docZip->getAttribute('NSU'),
                    'schema' => $schema,
                    'document_type' => $this->documentTypeFromSchema($schema),
                ];

                if ($includeXml) {
                    $document['xml'] = $this->decodeDocZip($encoded);
                }

                $result['documents'][] = $document;
            }
        }

        return $result;
    }

    /**
     * @return array<string, mixed>
     */
    public function mapDownloadResponse(string $xml): array
    {
        $mapped = $this->mapDistDFeResponse($xml, true);

        return [
            'cstat' => $mapped['cstat'],
            'xmotivo' => $mapped['xmotivo'],
            'tp_amb' => $mapped['tp_amb'],
            'ver_aplic' => $mapped['ver_aplic'],
            'dh_resp' => $mapped['dh_resp'],
            'document' => $mapped['documents'][0] ?? null,
            'raw_xml' => $xml,
        ];
    }

    /**
     * @return array<string, mixed>
     */
    public function mapConsultaChaveResponse(string $xml): array
    {
        $dom = $this->loadXml($xml);
        $ret = $dom->getElementsByTagName('retConsSitNFe')->item(0);

        if (! $ret instanceof DOMElement) {
            throw new RuntimeException('Resposta de consulta por chave inválida.');
        }

        $protocolo = null;
        $dhRecbto = null;

        $protNFe = $ret->getElementsByTagName('protNFe')->item(0);
        if ($protNFe instanceof DOMElement) {
            $protocolo = $this->nodeValue($protNFe, 'nProt');
            $dhRecbto = $this->nodeValue($protNFe, 'dhRecbto');
        }

        return [
            'cstat' => $this->nodeValue($ret, 'cStat'),
            'xmotivo' => $this->nodeValue($ret, 'xMotivo'),
            'chave' => $this->nodeValue($ret, 'chNFe'),
            'tp_amb' => $this->nodeValue($ret, 'tpAmb'),
            'protocolo' => $protocolo,
            'dh_recbto' => $dhRecbto,
            'raw_xml' => $xml,
        ];
    }

    private function loadXml(string $xml): DOMDocument
    {
        $dom = new DOMDocument;
        $internalErrors = libxml_use_internal_errors(true);

        try {
            if (! $dom->loadXML($xml)) {
                throw new RuntimeException('XML inválido retornado pela SEFAZ.');
            }
        } finally {
            libxml_clear_errors();
            libxml_use_internal_errors($internalErrors);
        }

        return $dom;
    }

    private function nodeValue(DOMElement $context, string $tagName): ?string
    {
        $node = $context->getElementsByTagName($tagName)->item(0);

        if (! $node) {
            return null;
        }

        $value = trim($node->textContent);

        return $value === '' ? null : $value;
    }

    private function decodeDocZip(string $encoded): ?string
    {
        if ($encoded === '') {
            return null;
        }

        $decoded = base64_decode($encoded, true);

        if ($decoded === false) {
            return null;
        }

        $content = @gzdecode($decoded);

        if ($content === false) {
            return null;
        }

        return $content;
    }

    private function documentTypeFromSchema(string $schema): string
    {
        if ($schema === '') {
            return 'unknown';
        }

        return strtolower(explode('_', $schema)[0]);
    }
}
